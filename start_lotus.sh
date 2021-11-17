cd ~
rm -rf .genesis-sectors/
rm -rf .lotus
rm -rf .lotusminer/
cd /opt/lotus/
rm -rf localnet.json
rm -rf devgen.car
export LOTUS_SKIP_GENESIS_CHECK=_yes_
./lotus fetch-params 8MiB
./lotus-seed pre-seal --sector-size 8MiB --num-sectors 1
./lotus-seed genesis new localnet.json
./lotus-seed genesis add-miner localnet.json ~/.genesis-sectors/pre-seal-t01000.json
./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false

