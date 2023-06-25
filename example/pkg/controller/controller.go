package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/morvencao/event-controller/pkg/mqtt"
)

type SyncReconciler struct {
	Cache *mqtt.Cache
	GVK   schema.GroupVersionKind
	Log   logr.Logger
}

func (r *SyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling instance", "namespace", req.Namespace, "name", req.Name)

	resourceMsg, err := r.Cache.Get(r.GVK, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	gvk := resourceMsg.Content.GroupVersionKind()

	if !resourceMsg.Content.GetDeletionTimestamp().IsZero() {
		// TODO: handle deletion
		// return ctrl.Result{}, c.handleDeletion(ctx, resourceMsg)

		r.Log.Info("deleting", "resource", gvk, "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, nil
	}

	// obj := resourceMsg.Content.DeepCopy()
	r.Log.Info("reconciling", "resource", gvk, "namespace", req.Namespace, "name", req.Name)

	// existingObj, err := c.reconcileObject(ctx, resourceMsg)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SyncReconciler) SetupWithManager(mgr ctrl.Manager, eventHub *mqtt.EventHub) error {
	eventClient := mqtt.NewEventClient()
	eventHub.Register(eventClient)
	targetGVKPredicate := func(object client.Object) bool {
		objGVK := object.GetObjectKind().GroupVersionKind()
		if objGVK == r.GVK {
			r.Log.Info(
				"predicate success",
				"gvk", objGVK.String(), "key", client.ObjectKeyFromObject(object))
			return true
		}
		r.Log.Info(
			"predicate filtered",
			"gvk", objGVK.String(), "key", client.ObjectKeyFromObject(object))
		return false
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("sync-controller").
		WatchesRawSource(&source.Channel{Source: eventClient.Channel},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.NewPredicateFuncs(targetGVKPredicate))).
		Complete(r)
}
