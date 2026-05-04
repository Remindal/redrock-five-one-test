module gateway

go 1.25.3

require (
	github.com/cloudwego/hertz v0.9.0
	github.com/cloudwego/kitex v0.10.0
	github.com/kitex-contrib/registry-etcd v0.2.0
)

replace (
	activity => ../activity
	seckill => ../seckill
	order => ../order
)