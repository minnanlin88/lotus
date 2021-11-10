package modules

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ipfs/go-graphsync"

	"github.com/filecoin-project/indexer-reference-provider/config"

	datatransfer "github.com/filecoin-project/go-data-transfer"

	"github.com/filecoin-project/lotus/node/repo"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/keytransform"

	dtimpl "github.com/filecoin-project/go-data-transfer/impl"
	dtnet "github.com/filecoin-project/go-data-transfer/network"
	dtgstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	provider "github.com/filecoin-project/indexer-reference-provider"
	"github.com/filecoin-project/indexer-reference-provider/engine"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"
)

func IndexerProvider(lc fx.Lifecycle, r repo.LockedRepo, h host.Host, ds dtypes.MetadataDS) (provider.Interface, error) {
	ipds := namespace.Wrap(ds, datastore.NewKey("/indexer-provider"))
	// TODO: supply reasonable defaults
	cfg := config.Ingest{
		LinkCacheSize:   0,
		LinkedChunkSize: 0,
		PubSubTopic:     "",
		PurgeLinkCache:  false,
	}
	// TODO:
	var privKey crypto.PrivKey
	var addrs []string

	dt, err := newIndexerProviderDataTransfer(lc, r, h, ipds)
	if err != nil {
		return nil, err
	}

	// TODO: Add a separate Start() method on the engine and call that from fx.Hook
	ctx := context.Background()
	e, err := engine.New(ctx, cfg, privKey, dt, h, ipds, addrs)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func newIndexerProviderDataTransfer(lc fx.Lifecycle, r repo.LockedRepo, h host.Host, ds datastore.Batching) (datatransfer.Manager, error) {
	net := dtnet.NewFromLibp2pHost(h)

	// TODO: initialize graphsync
	var gs graphsync.GraphExchange

	dtDs := namespace.Wrap(ds, datastore.NewKey("/datatransfer/transfers"))
	transport := dtgstransport.NewTransport(h.ID(), gs, net)
	dtPath := filepath.Join(r.Path(), "indexer-provider", "data-transfer")
	err := os.MkdirAll(dtPath, 0755) //nolint: gosec
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	dt, err := dtimpl.NewDataTransfer(dtDs, dtPath, net, transport)
	if err != nil {
		return nil, err
	}

	dt.OnReady(marketevents.ReadyLogger("indexer-provider data transfer"))
	lc.Append(fx.Hook{
		OnStart: dt.Start,
		OnStop:  dt.Stop,
	})
	return dt, nil
}
