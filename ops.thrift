
namespace go api

struct RequestAsgs {
    1: i64    id
}

struct RequestReply {
    1: bool     ret
}

struct ReleaseAsgs{
    1: i64   id
}

struct ReleaseReply{
    1: bool     ret
}

struct ReplyArgs {
    1: i64 id
}

struct ReplyReply {
    1: bool ret
}

struct ElectionAsgs{
    1: i64   id
}

struct ElectionReply{
    1: bool     ret
}
service Node_Operations {
    RequestReply    Request(1: RequestAsgs  req)
    ReleaseReply    Release(1: ReleaseAsgs  req)
    ReplyReply    Reply(1: ReplyArgs  req)
    ElectionReply        Election(1: ElectionAsgs  req)
}
