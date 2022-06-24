package logic

// 消费kafka的消息，依据消息类型进行判断
// 服务发现部分将依赖etcd
// 推送gateway部分将使用rpc的方式

// 1: 场景(在线)：
//   1.1: 用户点对点发送 -> 寻找gateway的节点 定点发送
//   1.2: 用户发送到房间(直播间) -> 广播gateway的所有节点，由gateway来控制是否维持和房间的链接来进行推送
