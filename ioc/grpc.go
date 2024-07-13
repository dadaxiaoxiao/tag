package ioc

import (
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/go-pkg/grpcx"
	"github.com/dadaxiaoxiao/go-pkg/grpcx/interceptors/logging"
	"github.com/dadaxiaoxiao/go-pkg/grpcx/interceptors/prometheus"
	ratelimit2 "github.com/dadaxiaoxiao/go-pkg/grpcx/interceptors/ratelimit"
	"github.com/dadaxiaoxiao/go-pkg/grpcx/interceptors/trace"
	pkgratelimit "github.com/dadaxiaoxiao/go-pkg/ratelimit"
	rgpc2 "github.com/dadaxiaoxiao/tag/internal/grpc"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"time"
)

func InitGRPCServer(ecli *clientv3.Client, tagRpc *rgpc2.TagServiceServer, redisClient redis.Cmdable, l accesslog.Logger) *grpcx.Server {
	type Config struct {
		Port    int   `yaml:"port"`
		EtcdTTL int64 `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	// 限流器
	limiter := pkgratelimit.NewRedisSlideWindowLimiter(redisClient,
		pkgratelimit.WithInterval(time.Second),
		pkgratelimit.WithRate(1000))

	// 这里添加拦截器
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// prometheus 拦截器
			prometheus.NewInterceptorBuilder("qinye", "demo_tag").BuildServer(),
			// 链路追踪
			trace.NewInterceptorBuilder(nil, nil).BuildServer(),
			// 日志拦截器
			logging.NewInterceptorBuilder(l).BuildServer(),
			// 服务器限流
			ratelimit2.NewInterceptorBuilder(limiter, "payment-server-limiter", l).BuildUnaryServerInterceptor(),
		))

	// 注册grpc
	tagRpc.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "tag",
		Log:        l,
		EtcdTTL:    cfg.EtcdTTL,
		EtcdClient: ecli,
	}
}
