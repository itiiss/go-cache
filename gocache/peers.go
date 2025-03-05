package gocache

type PeerPicker interface {
	// PickPeer 通过key得到相应节点的PeerGetter
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// Get 通过group和key得到缓存值
	Get(group, key string) ([]byte, error)
}
