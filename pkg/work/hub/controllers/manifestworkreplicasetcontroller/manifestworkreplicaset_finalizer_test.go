package manifestworkreplicasetcontroller

import (
	"context"
	"testing"
	"time"

	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fakeclient "open-cluster-management.io/api/client/work/clientset/versioned/fake"
	workinformers "open-cluster-management.io/api/client/work/informers/externalversions"
	"open-cluster-management.io/api/utils/work/v1/workapplier"

	helpertest "open-cluster-management.io/ocm/pkg/work/hub/test"
)

// Test finalize reconcile
func TestFinalizeReconcile(t *testing.T) {
	mwrSetTest := helpertest.CreateTestManifestWorkReplicaSet("mwrSet-test", "default", "place-test")
	mw, _ := CreateManifestWork(mwrSetTest, "cluster1", "plc1")
	fakeClient := fakeclient.NewSimpleClientset(mwrSetTest, mw)
	manifestWorkInformerFactory := workinformers.NewSharedInformerFactoryWithOptions(fakeClient, 1*time.Second)
	mwLister := manifestWorkInformerFactory.Work().V1().ManifestWorks().Lister()

	finalizerController := finalizeReconciler{
		workClient:         fakeClient,
		manifestWorkLister: mwLister,
		workApplier:        workapplier.NewWorkApplierWithTypedClient(fakeClient, mwLister),
	}

	// Set manifestWorkReplicaSet delete time AND Set finalizer
	timeNow := metav1.Now()
	mwrSetTest.DeletionTimestamp = &timeNow
	mwrSetTest.Finalizers = append(mwrSetTest.Finalizers, ManifestWorkReplicaSetFinalizer)

	_, _, err := finalizerController.reconcile(context.TODO(), mwrSetTest)
	if err != nil {
		t.Fatal(err)
	}

	updatetSet, err := fakeClient.WorkV1alpha1().ManifestWorkReplicaSets(mwrSetTest.Namespace).Get(context.TODO(), mwrSetTest.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Check mwrSetTest finalizer removed
	if slices.Contains(updatetSet.Finalizers, ManifestWorkReplicaSetFinalizer) {
		t.Fatal("Finalizer not deleted", mwrSetTest.Finalizers)
	}
}
