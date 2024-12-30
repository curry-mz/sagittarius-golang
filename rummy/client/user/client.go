package user

import (
	"context"
	"fmt"
	netHttp "net/http"

	"code.cd.local/sagittarius/sagittarius-golang/app/proxy"
	"code.cd.local/sagittarius/sagittarius-golang/cores/client/http"
	gErrors "code.cd.local/sagittarius/sagittarius-golang/cores/errors"

	"github.com/pkg/errors"
)

func GetAI(ctx context.Context, number int) ([]*AIData, error) {
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return nil, errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	req := map[string]interface{}{
		"count": number,
	}
	var rsp GetAIResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, getAIUrl), req, &rsp)
	if err != nil {
		return nil, errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return nil, errors.New(fmt.Sprintf("post %s, httpCode:%d", getAIUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return nil, gErrors.New(rsp.Status, rsp.Message)
	}
	return rsp.Data, nil
}

func GetUserInfos(ctx context.Context, userIds []int64) ([]*UserDetail, error) {
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return nil, errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	req := userIds
	var rsp GetUserInfosResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, getUsersUrl), req, &rsp)
	if err != nil {
		return nil, errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return nil, errors.New(fmt.Sprintf("post %s, httpCode:%d", getAIUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return nil, gErrors.New(rsp.Status, rsp.Message)
	}
	return rsp.Data, nil
}
