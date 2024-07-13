package ioc

import (
	"context"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/tag/internal/repository"
	"github.com/dadaxiaoxiao/tag/internal/repository/cache"
	"github.com/dadaxiaoxiao/tag/internal/repository/dao"
	"time"
)

// InitRepository 初始化 仓储层
func InitRepository(d dao.TagDAO, c cache.TagCache, l accesslog.Logger) repository.TagRepository {
	// 初始化
	repo := repository.NewTagRepository(d, c, l)
	go func() {
		// 执行缓存预加载
		// 或者启动的环境变量
		// 启动参数控制
		// 或者借助配置中心的开关
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer cancel()
		_ = repo.PreloadUserTags(ctx)
	}()
	return repo
}
