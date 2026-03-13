package service

import (
	"context"

	"github.com/go-logr/logr"
	examplev1 "github.com/ntlaletsi70/json-server-controller/api/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Writer struct {
	Client client.Client
	Log    logr.Logger
}

func NewWriter(c client.Client, log logr.Logger) *Writer {
	return &Writer{
		Client: c,
		Log:    log.WithName("jsonserver-status-writer"),
	}
}

func (w *Writer) Write(
	ctx context.Context,
	js *examplev1.JsonServer,
	err error,
) error {
	if err != nil {

		js.Status.State = examplev1.StateError
		js.Status.Message = err.Error()

	} else {

		js.Status.State = examplev1.StateSynced
		js.Status.Message = "Synced successfully"

	}

	w.Log.Info(
		"JsonServer status written",
		"name", js.Name,
		"namespace", js.Namespace,
		"state", js.Status.State,
	)

	return w.Client.Status().Update(ctx, js)
}
