namespace go order

struct CreateOrderReq {
    1: required i64 activity_id
    2: required string user_id
}

struct CreateOrderResp {
    1: required i32 code
    2: required string msg
    3: optional i64 order_id
}

struct QueryOrderReq {
    1: optional i64 order_id
    2: optional i64 activity_id
    3: optional string user_id
}

struct QueryOrderResp {
    1: required i32 code
    2: required string msg
    3: optional Order data
}

struct Order {
    1: required i64 id
    2: required i64 activity_id
    3: required string user_id
    4: required string status
    5: required string created_at
}

service OrderService {
    CreateOrderResp CreateOrder(1: CreateOrderReq req)
    QueryOrderResp QueryOrder(1: QueryOrderReq req)
}