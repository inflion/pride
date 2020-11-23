/*
Copyright 2020 inflion.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	instancev1beta1 "github.com/inflion/instance/api/v1beta1"
	inflionv1beta1 "github.com/inflion/pride/api/v1beta1"
)

// PrideReconciler reconciles a Pride object
type PrideReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inflion.inflion.com,resources=prides,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inflion.inflion.com,resources=prides/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inflion.inflion.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inflion.inflion.com,resources=instances/status,verbs=get;update;patch

func (r *PrideReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("pride", req.NamespacedName)

	var pride inflionv1beta1.Pride
	if err := r.Get(ctx, req.NamespacedName, &pride); err != nil {
		log.Error(err, "unable to get pride")
		return ctrl.Result{}, err
	}

	var instance instancev1beta1.Instance
	for _, instanceName := range pride.Spec.Instances {
		instanceNamespacedName := client.ObjectKey{
			Namespace: req.Namespace,
			Name:      instanceName,
		}

		err := r.Get(ctx, instanceNamespacedName, &instance)
		if err != nil {
			log.Error(err, "unable to get instance")
			return ctrl.Result{}, err
		}

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &instance, func() error {
			instance.Status.Sleeping = pride.Status.Sleeping
			return nil
		}); err != nil {
			log.Error(err, "unable to udpate instance")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PrideReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inflionv1beta1.Pride{}).
		Complete(r)
}
