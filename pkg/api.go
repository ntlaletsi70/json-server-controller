package service

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	examplev1 "github.com/ntlaletsi70/json-server-controller/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/utils/pointer"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Ensurer struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func NewEnsurer(c client.Client, scheme *runtime.Scheme, log logr.Logger) *Ensurer {
	return &Ensurer{
		Client: c,
		Scheme: scheme,
		Log:    log.WithName("Ensurer"),
	}
}

func (e *Ensurer) EnsureAll(ctx context.Context, js *examplev1.JsonServer) error {

	if err := e.ensureConfigMap(ctx, js); err != nil {
		return err
	}

	if err := e.ensureDeployment(ctx, js); err != nil {
		return err
	}

	if err := e.ensureService(ctx, js); err != nil {
		return err
	}

	return nil
}

func (e *Ensurer) ensureConfigMap(ctx context.Context, js *examplev1.JsonServer) error {

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      js.Name,
			Namespace: js.Namespace,
			Labels: map[string]string{
				"app": js.Name,
			},
		},
		Data: map[string]string{
			"db.json": normalizeJSON([]byte(js.Spec.JsonConfig)),
		},
	}

	if err := controllerutil.SetControllerReference(js, cm, e.Scheme); err != nil {
		return err
	}

	e.Log.Info("Ensuring ConfigMap", "name", js.Name, "jsonConfig", js.Spec.JsonConfig, "namespace", js.Namespace)

	return e.Client.Patch(
		ctx,
		cm,
		client.Apply,
		client.FieldOwner("JsonServer-controller"),
		client.ForceOwnership,
	)
}

func (e *Ensurer) ensureDeployment(ctx context.Context, js *examplev1.JsonServer) error {

	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      js.Name,
			Namespace: js.Namespace,
			Labels: map[string]string{
				"app": js.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &js.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": js.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": js.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "json-server",
							Image: "backplane/json-server",
							Args:  []string{"/data/db.json"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
									Name:          "http",
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "json-config",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "json-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: js.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(js, deploy, e.Scheme); err != nil {
		return err
	}

	e.Log.Info("Ensuring Deployment", "name", js.Name, "replicas", js.Spec.Replicas, "namespace", js.Namespace)

	return e.Client.Patch(
		ctx,
		deploy,
		client.Apply,
		&client.PatchOptions{
			FieldManager: "JsonServer-controller",
			Force:        pointer.Bool(true),
		},
	)
}

func (e *Ensurer) ensureService(ctx context.Context, js *examplev1.JsonServer) error {

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      js.Name,
			Namespace: js.Namespace,
			Labels: map[string]string{
				"app": js.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": js.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(3000),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(js, svc, e.Scheme); err != nil {
		return err
	}

	e.Log.Info("Ensuring Service", "name", js.Name, "", svc.Spec.Ports)

	return e.Client.Patch(
		ctx,
		svc,
		client.Apply,
		&client.PatchOptions{
			FieldManager: "JsonServer-controller",
			Force:        pointer.Bool(true),
		},
	)
}

func normalizeJSON(raw []byte) string {

	var obj interface{}

	if err := json.Unmarshal(raw, &obj); err != nil {
		return "{}"
	}

	switch obj.(type) {

	case []interface{}:
		// wrap array automatically
		wrapped := map[string]interface{}{
			"items": obj,
		}

		out, _ := json.MarshalIndent(wrapped, "", "  ")
		return string(out)

	default:
		return string(raw)
	}
}
