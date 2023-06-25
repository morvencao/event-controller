package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/morvencao/event-controller/example/pkg/controller"
	"github.com/morvencao/event-controller/pkg/mqtt"
)

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	setupLog := ctrl.Log.WithName("setup")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		LeaderElection:   false,
		LeaderElectionID: "80807133.example.com",
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cache := mqtt.NewCache()
	eventHub := mqtt.NewEventHub()

	decoder := mqtt.NewDecoder(mqtt.DecoderSinkList{cache, eventHub})
	conn := mqtt.NewConnection(
		ctrl.Log.WithName("mqtt-connection"), mqtt.ConnectionOptions{
			BrokerURLs: []string{"tcp://localhost:31320"},
			KeepAlive:  30 * time.Second,
			ClientID:   "client01",
			Username:   "admin",
			Password:   "password",
			Topic:      fmt.Sprintf("v1/client01/+/content"),
			OnMessage: func(m *paho.Publish) {
				if err := decoder.Decode(m); err != nil {
					panic(err)
				}
			},
		})

	// make sure event hub is started before mqtt connection
	if err := mgr.Add(eventHub); err != nil {
		setupLog.Error(err, "unable to set up event hub")
		os.Exit(1)
	}

	if err := mgr.Add(conn); err != nil {
		setupLog.Error(err, "unable to set up mqtt connection")
		os.Exit(1)
	}

	if err := (&controller.SyncReconciler{
		Cache: cache,
		GVK: schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		Log: ctrl.Log.WithName("sync-controller"),
	}).SetupWithManager(mgr, eventHub); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SyncReconciler")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
