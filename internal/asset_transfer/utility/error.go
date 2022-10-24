package utility_asset_transfer

import (
	"errors"
	"fmt"
	sig_graph_grpc "sig_graph_scp/internal/grpc"
	"sig_graph_scp/pkg/utility"
)

var ErrPeerGeneralError = errors.New("peer general error")
var ErrUnhandledPeerGrpcError = errors.New("unhandled error")

func WrapGrpcError(err *sig_graph_grpc.Error) error {
	code := err.GetCode()
	switch code {
	case sig_graph_grpc.ErrorCode_SUCCESS:
		return nil
	case sig_graph_grpc.ErrorCode_ALREADY_EXISTS:
		return fmt.Errorf("%w: %s", utility.ErrAlreadyExists, err.ErrorMessage)
	case sig_graph_grpc.ErrorCode_INVALID_ARGUMENT:
		return fmt.Errorf("%w: %s", utility.ErrInvalidArgument, err.ErrorMessage)
	case sig_graph_grpc.ErrorCode_NOT_FOUND:
		return fmt.Errorf("%w: %s", utility.ErrNotFound, err.ErrorMessage)
	case sig_graph_grpc.ErrorCode_GENERAL_ERROR:
		return fmt.Errorf("%w: %s", ErrPeerGeneralError, err.ErrorMessage)

	}

	// shouldn't really go here
	return fmt.Errorf("%w: code - %d; message - %s", ErrPeerGeneralError, code, err.GetErrorMessage())
}

func ToGrpcError(err error) *sig_graph_grpc.Error {
	switch err {
	case nil:
		return &sig_graph_grpc.Error{
			Code:         sig_graph_grpc.ErrorCode_SUCCESS,
			ErrorMessage: "success",
		}
	case utility.ErrNotFound:
		return &sig_graph_grpc.Error{
			Code:         sig_graph_grpc.ErrorCode_NOT_FOUND,
			ErrorMessage: err.Error(),
		}
	case utility.ErrAlreadyExists:
		return &sig_graph_grpc.Error{
			Code:         sig_graph_grpc.ErrorCode_ALREADY_EXISTS,
			ErrorMessage: err.Error(),
		}
	case utility.ErrInvalidArgument:
		return &sig_graph_grpc.Error{
			Code:         sig_graph_grpc.ErrorCode_INVALID_ARGUMENT,
			ErrorMessage: err.Error(),
		}
	default:
		return &sig_graph_grpc.Error{
			Code:         sig_graph_grpc.ErrorCode_GENERAL_ERROR,
			ErrorMessage: err.Error(),
		}
	}
}
