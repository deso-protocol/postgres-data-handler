module github.com/deso-protocol/postgres-data-handler

go 1.23.0

toolchain go1.23.2

replace github.com/deso-protocol/core => ../core/

replace github.com/deso-protocol/backend => ../backend/

replace github.com/deso-protocol/state-consumer => ../state-consumer/

require (
	github.com/deso-protocol/core v0.0.0-00010101000000-000000000000
	github.com/deso-protocol/state-consumer v0.0.0-00010101000000-000000000000
	github.com/golang/glog v1.2.2
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.19.0
	github.com/uptrace/bun v1.1.17
	github.com/uptrace/bun/dialect/pgdialect v1.1.14
	github.com/uptrace/bun/driver/pgdriver v1.1.14
	github.com/uptrace/bun/extra/bunbig v1.1.14
	github.com/uptrace/bun/extra/bundebug v1.1.14
	gopkg.in/DataDog/dd-trace-go.v1 v1.68.0
)

require (
	cloud.google.com/go v0.112.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/storage v1.38.0 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/btcsuite/btcd v0.24.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btclog v0.0.0-20241003133417-09c4e92e319c // indirect
	github.com/btcsuite/btcutil v1.0.2 // indirect
	github.com/bwesterb/go-ristretto v1.2.3 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.4.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/decred/dcrd/lru v1.1.3 // indirect
	github.com/deso-protocol/go-deadlock v1.0.0 // indirect
	github.com/deso-protocol/go-merkle-tree v1.0.0 // indirect
	github.com/dgraph-io/badger/v3 v3.2103.5 // indirect
	github.com/dgraph-io/ristretto v1.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/ethereum/go-ethereum v1.14.11 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gernest/mention v2.0.0+incompatible // indirect
	github.com/git-chglog/git-chglog v0.15.4 // indirect
	github.com/go-co-op/gocron/v2 v2.11.0 // indirect
	github.com/go-pg/pg/v10 v10.13.0 // indirect
	github.com/go-pg/zerochecker v0.2.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20241008150032-332c0e1a4a34 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/h2non/bimg v1.1.5 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/go-types v0.0.0-20210723172823-2deba1f80ba7 // indirect
	github.com/kevinburke/rest v0.0.0-20210506044642-5611499aa33c // indirect
	github.com/kevinburke/twilio-go v0.0.0-20210327194925-1623146bcf73 // indirect
	github.com/klauspost/compress v1.17.10 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/goveralls v0.0.12 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/nyaruka/phonenumbers v1.0.69 // indirect
	github.com/oleiade/lane v1.0.1 // indirect
	github.com/onflow/crypto v0.25.2 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/robinjoseph08/go-pg-migrations/v3 v3.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sendgrid/rest v2.6.4+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.10.0+incompatible // indirect
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tinylib/msgp v1.2.2 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/tsuyoshiwada/go-gitcmd v0.0.0-20180205145712-5f1f5f9475df // indirect
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.2.1 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/unrolled/secure v1.16.0 // indirect
	github.com/urfave/cli v1.22.1 // indirect
	github.com/vmihailenco/bufpool v0.1.11 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20241004190924-225e2abe05e6 // indirect
	golang.org/x/image v0.15.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/api v0.171.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.7 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/kyokomi/emoji.v1 v1.5.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mellium.im/sasl v0.3.2 // indirect
)
