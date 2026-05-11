namespace go activity

struct CreateActivityReq {
    1: required string name
    2: required i32 stock
    3: required string start_time
    4: required string end_time
}

struct CreateActivityResp {
    1: required i32 code
    2: required string msg
    3: optional i64 activity_id
}

struct GetActivityReq {
    1: required i64 activity_id
}

struct GetActivityResp {
    1: required i32 code
    2: required string msg
    3: optional Activity data
}

struct Activity {
    1: required i64 id
    2: required string name
    3: required i32 stock
    4: required i32 remain_stock
    5: required string start_time
    6: required string end_time
    7: required i32 status
}

service ActivityService {
    CreateActivityResp CreateActivity(1: CreateActivityReq req)
    GetActivityResp GetActivity(1: GetActivityReq req)
}