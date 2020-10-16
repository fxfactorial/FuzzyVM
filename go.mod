module github.com/MariusVanDerWijden/FuzzyVM

go 1.15

require (
	github.com/ethereum/go-ethereum v1.9.23
	github.com/holiman/goevmlab v0.0.0-20200925112252-8249743488ae
	github.com/korovkin/limiter v0.0.0-20190919045942-dac5a6b2a536
	github.com/pkg/errors v0.9.1
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace github.com/MariusVanDerWijden/FuzzyVM/filler => ./filler

replace github.com/ethereum/go-ethereum => /home/matematik/go/src/github.com/ethereum/go-ethereum

replace github.com/holiman/goevmlab => /home/matematik/ethereum/goevmlab
