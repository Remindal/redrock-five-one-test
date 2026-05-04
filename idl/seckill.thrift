namespace go seckill

struct DoSeckillReq {
    1: required i64 activity_id
    2: required string user_id
}

struct DoSeckillResp {
    1: required i32 code
    2: required string msg
    3: optional string order_id
    4: optional string status
}

service SeckillService {
    DoSeckillResp DoSeckill(1: DoSeckillReq req)
}