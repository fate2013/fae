package engine

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/funkygao/fae/servant/gen-go/fun/rpc"
)

func GetClient(serverAddr string) (*rpc.FunServantClient, thrift.TTransport) {
	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocket(serverAddr)
	if err != nil {
		panic(err)
	}

	useTransport := transportFactory.GetTransport(transport)
	client := rpc.NewFunServantClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		panic(err)
	}

	return client, transport
}
