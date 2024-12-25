
namespace go api

struct RequestAsgs {
    1: i64    consumer
    2: string port
}

struct RequestReply {
    1: bool     ret
}

struct ReleaseAsgs{
    1: i64   consumer
}

struct ReleaseReply{
    1: bool     ret
}

service Client_Operations {
    RequestReply    Request(1: RequestAsgs  req)
    ReleaseReply    Release(1: ReleaseAsgs  req)
}

struct ReplyArgs {
    1: bool resource
}

struct ReplyReply {
    1: bool ret
}

service Server_Operations {
    ReplyReply    Reply(1: ReplyArgs  req)
}
