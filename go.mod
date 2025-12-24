module epaxos

go 1.25.4

replace (
	code.google.com/p/gogoprotobuf => github.com/gogo/protobuf v1.3.2
	code.google.com/p/leveldb-go/leveldb => github.com/syndtr/goleveldb v1.0.0
)

require (
	code.google.com/p/leveldb-go/leveldb v0.0.0-00010101000000-000000000000 // indirect
	github.com/go-distributed/epaxos v0.0.0-20151126153908-f68e77277f2d // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/leveldb v0.0.0-20170107010102-259d9253d719 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/stretchr/testify v1.11.1 // indirect
)

replace github.com/go-distributed/epaxos => ../epaxos
