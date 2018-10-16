package driver

import (
	_ "github.com/yatlabs/driver/logger-default"
	_ "github.com/yatlabs/driver/logger-file"
	_ "github.com/yatlabs/driver/mutex-default"
	_ "github.com/yatlabs/driver/mutex-memcache"
	_ "github.com/yatlabs/driver/mutex-redis"
	_ "github.com/yatlabs/driver/session-default"
	_ "github.com/yatlabs/driver/session-redis"
	_ "github.com/yatlabs/driver/session-memcache"
	_ "github.com/yatlabs/driver/session-file"
	_ "github.com/yatlabs/driver/session-memory"

	_ "github.com/yatlabs/driver/cache-default"
	_ "github.com/yatlabs/driver/cache-file"
	_ "github.com/yatlabs/driver/cache-memory"
	_ "github.com/yatlabs/driver/cache-memcache"
	_ "github.com/yatlabs/driver/cache-redis"
	_ "github.com/yatlabs/driver/data-postgres"
	_ "github.com/yatlabs/driver/data-cockroach"
	_ "github.com/yatlabs/driver/file-default"
	
	_ "github.com/yatlabs/driver/plan-default"
	_ "github.com/yatlabs/driver/event-default"
	_ "github.com/yatlabs/driver/event-redis"
	_ "github.com/yatlabs/driver/queue-default"
	_ "github.com/yatlabs/driver/queue-redis"

	_ "github.com/yatlabs/driver/http-default"
	_ "github.com/yatlabs/driver/view-default"
	_ "github.com/yatlabs/driver/socket-default"

)
