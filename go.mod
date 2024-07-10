module github.com/wdvxdr1123/ZeroBot

go 1.19

require (
	github.com/FloatTech/ttl v0.0.0-20220715042055-15612be72f5b
	github.com/RomiChan/syncx v0.0.0-20240418144900-b7402ffdebc7
	github.com/RomiChan/websocket v1.4.3-0.20220227141055-9b2c6168c9c5
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/t-tomalak/logrus-easy-formatter v0.0.0-20190827215021-c074f06c5816
	github.com/tidwall/gjson v1.14.4
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	modernc.org/libc v1.21.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.4.0 // indirect
	modernc.org/sqlite v1.20.0 // indirect
)

require (
	github.com/FloatTech/sqlite v1.6.3
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace modernc.org/sqlite => github.com/fumiama/sqlite3 v1.20.0-with-win386

replace github.com/remyoudompheng/bigfft => github.com/fumiama/bigfft v0.0.0-20211011143303-6e0bfa3c836b
