package service

import (
	"context"

	examplev1 "github.com/ntlaletsi70/json-server-controller/api/v1"
)

type JsonServerService struct {
	status  *Writer
	ensurer *Ensurer
}

func NewJsonServerService(
	status *Writer,
	ensurer *Ensurer,
) *JsonServerService {

	return &JsonServerService{
		status:  status,
		ensurer: ensurer,
	}
}

func (s *JsonServerService) Reconcile(
	ctx context.Context,
	js *examplev1.JsonServer,
) error {

	//------------------------------------------------
	// Validate JSON
	//------------------------------------------------

	if err := validateJSONPayload(js.Spec.JsonConfig); err != nil {
		return s.status.Write(ctx, js, err)
	}

	//------------------------------------------------
	// Ensure resources
	//------------------------------------------------

	err := s.ensurer.EnsureAll(ctx, js)

	return s.status.Write(ctx, js, err)
}
