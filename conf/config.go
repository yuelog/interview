package conf

type Config struct {
	BufferSize int // 缓冲通道大小
	WorkerNum  int // 业务处理协程池大小        10

	Iv     string // 加密
	Cipher string // 加密

	// online model
	IdGenRedisPrefix     string // IdGen的redis前缀
	IdGenHost            string // host
	IdGenPassword        string // redis pass
	IdGenDatabase        int    // 数据库
	IdGenPoolMaxIdle     int    // 最大空闲连接数
	IdGenPoolMaxActive   int    // 一个pool支持的最大连接数，设置为0表示无限制
	IdGenPoolIdleTimeout int    // 空闲连接超时时间，超过超时时间的空闲连接会被关闭，为0表示空闲连接不会被关闭
	IdGenPoolWait        bool   // pool Get()方法是否会被阻塞
	IdGenConnectTimeout  int    // redis建立连接的超时时间
	IdGenReadTimeout     int    // redis读取数据的超时时间
	IdGenWriteTimeout    int    // redis写入数据的超时时间

	// 数据redis
	OnlineRedisPrefix     string // IdGen的redis前缀
	OnlineHost            string // host
	OnlinePassword        string // redis pass
	OnlineDatabase        int    // 数据库
	OnlinePoolMaxIdle     int    // 最大空闲连接数
	OnlinePoolMaxActive   int    // 一个pool支持的最大连接数，设置为0表示无限制
	OnlinePoolIdleTimeout int    // 空闲连接超时时间，超过超时时间的空闲连接会被关闭，为0表示空闲连接不会被关闭
	OnlinePoolWait        bool   // pool Get()方法是否会被阻塞
	OnlineConnectTimeout  int    // redis建立连接的超时时间
	OnlineReadTimeout     int    // redis读取数据的超时时间
	OnlineWriteTimeout    int    // redis写入数据的超时时间

	SqlDriverNameMaster     string // sql数据库类型 (master)
	SqlDataSourceNameMaster string // 数据库资源字符串 (master)

	SqlDriverNameSlave     string // sql数据库类型 (slave)
	SqlDataSourceNameSlave string // 数据库资源字符串 (slave)

	UcUrl      string // uc接口host
	UcBusiness string // uc接口使用的business
	UcToken    string // uc接口使用的token

	OldImNameArr string // 老IM接收服务    注册的服务名称字符串，多个用逗号隔开

	Sharding map[string]int //分表配置
}

var (
	config Config
)

func GetIns() *Config {
	return &config
}

func (c *Config) CheckDefault() {

}
