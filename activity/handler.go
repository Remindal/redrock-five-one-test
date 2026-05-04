package main

import (
	"activity/biz/service"
	"activity/kitex_gen/activity"
	"context"
)

type ActivityServiceImpl struct{}

func (s *ActivityServiceImpl) CreateActivity(ctx context.Context, req *activity.CreateActivityReq) (resp *activity.CreateActivityResp, err error) {
	return service.NewActivityService(ctx).CreateActivity(req)
}

func (s *ActivityServiceImpl) GetActivity(ctx context.Context, req *activity.GetActivityReq) (resp *activity.GetActivityResp, err error) {
	return service.NewActivityService(ctx).GetActivity(req)
}
