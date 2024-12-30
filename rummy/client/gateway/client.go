package gateway

import (
	"context"
	"fmt"
	netHttp "net/http"

	"code.cd.local/sagittarius/sagittarius-golang/app/proxy"
	"code.cd.local/sagittarius/sagittarius-golang/cores/client/http"
	gErrors "code.cd.local/sagittarius/sagittarius-golang/cores/errors"

	"github.com/pkg/errors"
)

func PushMessage(ctx context.Context, req *PushMessageRequest) error {
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	var rsp PushMessageResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, PushMessageUrl), req, &rsp)
	if err != nil {
		return errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return errors.New(fmt.Sprintf("post %s, httpCode:%d", PushMessageUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return gErrors.New(rsp.Status, rsp.Message)
	}
	return nil
}
