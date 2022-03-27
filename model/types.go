package model

import "database/sql"

// 获取不到代理信息，默认归属root下，为直客
type MemberData struct {
	Member
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	IsRisk   int    `json:"is_risk"`
}
type MemberPageData struct {
	T int64        `json:"t"`
	D []MemberData `json:"d"`
	S uint         `json:"s"`
}

type Member struct {
	UID                string `db:"uid" json:"uid,omitempty"`
	Username           string `db:"username" json:"username,omitempty"`                         //会员名
	Password           string `db:"password" json:"password,omitempty"`                         //密码
	RealnameHash       uint64 `db:"realname_hash" json:"realname_hash,omitempty"`               //真实姓名哈希
	EmailHash          uint64 `db:"email_hash" json:"email_hash,omitempty"`                     //邮件地址哈希
	PhoneHash          uint64 `db:"phone_hash" json:"phone_hash,omitempty"`                     //电话号码哈希
	Prefix             string `db:"prefix" json:"prefix,omitempty"`                             //站点前缀
	WithdrawPwd        uint64 `db:"withdraw_pwd" json:"withdraw_pwd,omitempty"`                 //取款密码哈希
	Regip              string `db:"regip" json:"regip,omitempty"`                               //注册IP
	RegUrl             string `db:"reg_url" json:"reg_url,omitempty"`                           //注册域名
	RegDevice          string `db:"reg_device" json:"reg_device,omitempty"`                     //注册设备号
	CreatedAt          uint32 `db:"created_at" json:"created_at,omitempty"`                     //注册时间
	LastLoginIp        string `db:"last_login_ip" json:"last_login_ip,omitempty"`               //最后登陆ip
	LastLoginAt        uint32 `db:"last_login_at" json:"last_login_at,omitempty"`               //最后登陆时间
	SourceId           uint8  `db:"source_id" json:"source_id,omitempty"`                       //注册来源 1 pc 2h5 3 app
	FirstDepositAt     uint32 `db:"first_deposit_at" json:"first_deposit_at,omitempty"`         //首充时间
	FirstDepositAmount string `db:"first_deposit_amount" json:"first_deposit_amount,omitempty"` //首充金额
	FirstBetAt         uint32 `db:"first_bet_at" json:"first_bet_at,omitempty"`                 //首投时间
	FirstBetAmount     string `db:"first_bet_amount" json:"first_bet_amount,omitempty"`         //首投金额
	TopUid             string `db:"top_uid" json:"top_uid,omitempty"`                           //总代uid
	TopName            string `db:"top_name" json:"top_name,omitempty"`                         //总代代理
	ParentUid          string `db:"parent_uid" json:"parent_uid,omitempty"`                     //上级uid
	ParentName         string `db:"parent_name" json:"parent_name,omitempty"`                   //上级代理
	BankcardTotal      uint8  `db:"bankcard_total" json:"bankcard_total,omitempty"`             //用户绑定银行卡的数量
	LastLoginDevice    string `db:"last_login_device" json:"last_login_device,omitempty"`       //最后登陆设备
	LastLoginSource    uint8  `db:"last_login_source" json:"last_login_source,omitempty"`       //上次登录设备来源:1=pc,2=h5,3=ios,4=andriod
	Remarks            string `db:"remarks" json:"remarks,omitempty"`                           //备注
	State              uint8  `db:"state" json:"state,omitempty"`                               //状态 1正常 2禁用
	Balance            string `db:"balance" json:"balance,omitempty"`                           //余额
	LockAmount         string `db:"lock_amount" json:"lock_amount,omitempty"`                   //锁定金额
	Commission         string `db:"commission" json:"commission,omitempty"`                     //佣金
	MaintainName       string `db:"maintain_name" json:"maintain_name"`                         //维护人
}

// MemberPlatform 会员场馆表
type MemberPlatform struct {
	ID                    string `db:"id" json:"id" redis:"id"`                                                                //
	Username              string `db:"username" json:"username" redis:"username"`                                              //用户名
	Pid                   string `db:"pid" json:"pid" redis:"pid"`                                                             //场馆ID
	Password              string `db:"password" json:"password" redis:"password"`                                              //平台密码
	Balance               string `db:"balance" json:"balance" redis:"balance"`                                                 //平台余额
	State                 int    `db:"state" json:"state" redis:"state"`                                                       //状态:1=正常,2=锁定
	CreatedAt             uint32 `db:"created_at" json:"created_at" redis:"created_at"`                                        //
	TransferIn            int    `db:"transfer_in" json:"transfer_in" redis:"transfer_in"`                                     //0:没有转入记录1:有
	TransferInProcessing  int    `db:"transfer_in_processing" json:"transfer_in_processing" redis:"transfer_in_processing"`    //0:没有转入等待记录1:有
	TransferOut           int    `db:"transfer_out" json:"transfer_out" redis:"transfer_out"`                                  //0:没有转出记录1:有
	TransferOutProcessing int    `db:"transfer_out_processing" json:"transfer_out_processing" redis:"transfer_out_processing"` //0:没有转出等待记录1:有
	Extend                uint64 `db:"extend" json:"extend" redis:"extend"`                                                    //兼容evo
}

type MBBalance struct {
	UID        string `db:"uid" json:"uid"`
	Balance    string `db:"balance" json:"balance"`         //余额
	LockAmount string `db:"lock_amount" json:"lock_amount"` //锁定额度
	Commission string `db:"commission" json:"commission"`   //代理余额
}

//账变表
type MemberTransaction struct {
	AfterAmount  string `db:"after_amount"`  //账变后的金额
	Amount       string `db:"amount"`        //用户填写的转换金额
	BeforeAmount string `db:"before_amount"` //账变前的金额
	BillNo       string `db:"bill_no"`       //转账|充值|提现ID
	CashType     int    `db:"cash_type"`     //0:转入1:转出2:转入失败补回3:转出失败扣除4:存款5:提现
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
	Prefix       string `db:"prefix"`        //站点前缀
}

//场馆转账表
type MemberTransfer struct {
	AfterAmount  string `db:"after_amount"`  //转账后的金额
	Amount       string `db:"amount"`        //金额
	Automatic    int    `db:"automatic"`     //1:自动转账2:脚本确认3:人工确认
	BeforeAmount string `db:"before_amount"` //转账前的金额
	BillNo       string `db:"bill_no"`       //
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	PlatformID   string `db:"platform_id"`   //三方场馆ID
	State        int    `db:"state"`         //0:失败1:成功2:处理中3:脚本确认中4:人工确认中
	TransferType int    `db:"transfer_type"` //0:转入1:转出
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
	ConfirmAt    int64  `db:"confirm_at"`    //确认时间
	ConfirmUid   uint64 `db:"confirm_uid"`   //确认人uid
	ConfirmName  string `db:"confirm_name"`  //确认人名
}

type CommissionTransferData struct {
	S int                  `json:"s"`
	D []CommissionTransfer `json:"d"`
	T int64                `json:"t"`
}

type CommissionsData struct {
	S int           `json:"s"`
	D []Commissions `json:"d"`
	T int64         `json:"t"`
}

type CommissionTransfer struct {
	ID           string `json:"id" db:"id"`
	UID          string `json:"uid" db:"uid"`                     //用户ID
	Username     string `json:"username" db:"username"`           //用户名
	ReceiveUID   string `json:"receive_uid" db:"receive_uid"`     //用户ID
	ReceiveName  string `json:"receive_name" db:"receive_name"`   //用户名
	TransferType int    `json:"transfer_type" db:"transfer_type"` //转账类型 2 佣金提取 3佣金下发
	Amount       string `json:"amount" db:"amount"`               //金额
	CreatedAt    int64  `json:"created_at" db:"created_at"`       //创建时间
	State        int    `json:"state" db:"state"`                 //1 审核中 2 审核通过 3 审核不通过
	Automatic    int    `json:"automatic" db:"automatic"`         // 1自动 2手动
	ReviewAt     int64  `json:"review_at" db:"review_at"`         //审核时间
	ReviewUid    string `json:"review_uid" db:"review_uid"`       //审核人uid
	ReviewName   string `json:"review_name" db:"review_name"`     //审核人名
	ReviewRemark string `json:"review_remark" db:"review_remark"` //审核备注
	Prefix       string `json:"prefix" db:"prefix"`
}

type MemberLoginLogData struct {
	S int              `json:"s"`
	D []MemberLoginLog `json:"d"`
	T int64            `json:"t"`
}

type MemberLoginLog struct {
	Username string `msg:"username" json:"username"`
	IP       int64  `msg:"ip" json:"ip"`
	IPS      string `msg:"ips" json:"ips"`
	Device   string `msg:"device" json:"device"`
	DeviceNo string `msg:"device_no" json:"device_no"`
	Date     uint32 `msg:"date" json:"date"`
	Serial   string `msg:"serial" json:"serial"`
	Agency   bool   `msg:"agency" json:"agency"`
	Parents  string `msg:"parents" json:"parents"`
	IsRisk   int    `msg:"-" json:"is_risk"`
}

type memberRemarkLogData struct {
	S int                `json:"s"`
	D []MemberRemarksLog `json:"d"`
	T int64              `json:"t"`
}

// 用户备注日志
type MemberRemarksLog struct {
	ID        string `msg:"id" json:"id"`
	UID       string `msg:"uid" json:"uid"`
	Username  string `msg:"username" json:"username"`
	Msg       string `msg:"msg" json:"msg"`
	File      string `msg:"file" json:"file"`
	AdminName string `msg:"admin_name" json:"admin_name"`
	CreatedAt int64  `msg:"created_at" json:"created_at"`
	Prefix    string `msg:"prefix" json:"prefix"`
}

// MemberAdjust db structure
type MemberAdjust struct {
	ID            string  `db:"id" json:"id"`
	UID           string  `db:"uid" json:"uid"` // 会员id
	Prefix        string  `db:"prefix" json:"prefix"`
	Ty            int     `db:"ty" json:"ty"`                             //来源
	Username      string  `db:"username" json:"username"`                 // 会员username
	TopUid        string  `db:"top_uid" json:"top_uid,omitempty"`         //总代uid
	TopName       string  `db:"top_name" json:"top_name,omitempty"`       //总代代理
	ParentUid     string  `db:"parent_uid" json:"parent_uid,omitempty"`   //上级uid
	ParentName    string  `db:"parent_name" json:"parent_name,omitempty"` //上级代理
	Amount        float64 `db:"amount" json:"amount"`                     // 调整金额
	AdjustType    int     `db:"adjust_type" json:"adjust_type"`           // 调整类型:1=系统调整,2=输赢调整,3=线下转卡充值
	AdjustMode    int     `db:"adjust_mode" json:"adjust_mode"`           // 调整方式:1=上分,2=下分
	IsTurnover    int     `db:"is_turnover" json:"is_turnover"`           // 是否需要流水限制:1=需要,0=不需要
	TurnoverMulti int     `db:"turnover_multi" json:"turnover_multi"`     // 流水倍数
	ApplyRemark   string  `db:"apply_remark" json:"apply_remark"`         // 申请备注
	ReviewRemark  string  `db:"review_remark" json:"review_remark"`       // 审核备注
	State         int     `db:"state" json:"state"`                       // 状态:1=审核中,2=审核通过,3=审核未通过
	HandOutState  int     `db:"hand_out_state" json:"hand_out_state"`     // 上下分状态 1 失败 2成功 3场馆上分处理中
	Images        string  `db:"images" json:"images"`                     // 图片地址
	ApplyAt       int64   `db:"apply_at" json:"apply_at"`                 // 申请时间
	ApplyUid      string  `db:"apply_uid" json:"apply_uid"`               // 申请人uid
	ApplyName     string  `db:"apply_name" json:"apply_name"`             // 申请人
	ReviewAt      int64   `db:"review_at" json:"review_at"`               // 审核时间
	ReviewUid     string  `db:"review_uid" json:"review_uid"`             // 审核人uid
	ReviewName    string  `db:"review_name" json:"review_name"`           // 审核人
	IsRisk        int     `db:"-" json:"is_risk"`
}

type BankcardData struct {
	BankCard
	RealName string `json:"realname" name:"realname"`
	Bankcard string `json:"bankcard_no" name:"bankcard"`
}

type BankCard struct {
	ID           string `db:"id" json:"id"`
	UID          string `db:"uid" json:"uid"`
	Username     string `db:"username" json:"username"`
	BankAddress  string `db:"bank_address" json:"bank_address"`
	BankID       string `db:"bank_id" json:"bank_id"`
	BankBranch   string `db:"bank_branch_name" json:"bank_branch_name"`
	State        int    `db:"state" json:"state"`
	BankcardHash string `db:"bank_card_hash" json:"bank_card_hash"`
	CreatedAt    uint64 `db:"created_at" json:"created_at"`
	Prefix       string `db:"prefix" json:"prefix"`
}

type DividendData struct {
	T   int64             `json:"t"`
	D   []MemberDividend  `json:"d"`
	Agg map[string]string `json:"agg"`
}

type MemberDividend struct {
	ID            string  `db:"id" json:"id"`
	UID           string  `db:"uid" json:"uid"`
	Prefix        string  `db:"prefix" json:"prefix"`
	Wallet        int     `db:"wallet" json:"wallet"`
	Ty            int     `db:"ty" json:"ty"`
	WaterLimit    uint8   `db:"water_limit" json:"water_limit"`
	PlatformID    string  `db:"platform_id" json:"platform_id"`
	Username      string  `db:"username" json:"username"`
	TopUid        string  `db:"top_uid" json:"top_uid,omitempty"`         //总代uid
	TopName       string  `db:"top_name" json:"top_name,omitempty"`       //总代代理
	ParentUid     string  `db:"parent_uid" json:"parent_uid,omitempty"`   //上级uid
	ParentName    string  `db:"parent_name" json:"parent_name,omitempty"` //上级代理
	Amount        float64 `db:"amount" json:"amount"`
	HandOutAmount float64 `db:"hand_out_amount" json:"hand_out_amount"`
	WaterFlow     float64 `db:"water_flow" json:"water_flow"`
	State         int     `db:"state" json:"state"`
	HandOutState  int     `db:"hand_out_state" json:"hand_out_state"`
	Automatic     int     `db:"automatic" json:"automatic"`
	Remark        string  `db:"remark" json:"remark"`
	ReviewRemark  string  `db:"review_remark" json:"review_remark"`
	ApplyAt       uint64  `db:"apply_at" json:"apply_at"`
	ApplyUid      string  `db:"apply_uid" json:"apply_uid"`
	ApplyName     string  `db:"apply_name" json:"apply_name"`
	ReviewAt      uint64  `db:"review_at" json:"review_at"`
	ReviewUid     string  `db:"review_uid" json:"review_uid"`
	ReviewName    string  `db:"review_name" json:"review_name"`
	IsRisk        int     `db:"-" json:"is_risk"`
}

type BannerData struct {
	T int64    `json:"t"`
	D []Banner `json:"d"`
	S uint     `json:"s"`
}

type Banner struct {
	ID          string `json:"id" db:"id" rule:"none"`                                                                       //
	Title       string `json:"title" db:"title" msg:"title error" rule:"filter" name:"title"`                                //标题
	Device      string `json:"device" db:"device" rule:"sDigit" msg:"device error" name:"device"`                            //设备类型(1,2)
	RedirectURL string `json:"redirect_url" db:"redirect_url" rule:"none" msg:"redirect_url error" name:"redirect_url"`      //跳转地址
	Images      string `json:"images" db:"images" rule:"none"`                                                               //图片路径
	Seq         string `json:"seq" db:"seq" rule:"digit" min:"1" max:"100" msg:"seq error" name:"seq"`                       //排序
	Flags       string `json:"flags" db:"flags" rule:"digit" min:"1" max:"10" msg:"flags error" name:"flags"`                //广告类型
	ShowType    string `json:"show_type" db:"show_type" rule:"digit" min:"1" max:"2" msg:"show_type error" name:"show_type"` //1 永久有效 2 指定时间
	ShowAt      string `json:"show_at" db:"show_at" rule:"none" msg:"show_at error" name:"show_at"`                          //开始展示时间
	HideAt      string `json:"hide_at" db:"hide_at" rule:"none" msg:"hide_at error" name:"hide_at"`                          //结束展示时间
	URLType     string `json:"url_type" db:"url_type" rule:"digit" min:"0" max:"3" msg:"url_type error" name:"url_type"`     //链接类型 1站内 2站外
	UpdatedName string `json:"updated_name" db:"updated_name" rule:"none"`                                                   //更新人name
	UpdatedUID  string `json:"updated_uid" db:"updated_uid" rule:"none"`                                                     //更新人id
	UpdatedAt   string `json:"updated_at" db:"updated_at" rule:"none"`                                                       //更新时间
	State       uint8  `json:"state" db:"state" rule:"none"`                                                                 //0:关闭1:开启
	Prefix      string `json:"prefix" db:"prefix" rule:"none"`
}

type BlacklistData struct {
	T int64       `json:"t"`
	D []Blacklist `json:"d"`
	S uint        `json:"s"`
}

type Blacklist struct {
	ID          string `json:"id" db:"id"`                                 //id
	Ty          int    `json:"ty" db:"ty"`                                 //黑名单类型
	Value       string `json:"value" db:"value"`                           //黑名单类型值
	Accounts    string `json:"accounts" db:"accounts"`                     //关联账号名 逗号分割
	Remark      string `json:"remark" db:"remark"`                         //备注
	CreatedAt   string `json:"created_at" db:"created_at" rule:"none"`     //添加时间
	CreatedUID  string `json:"created_uid" db:"created_uid" rule:"none"`   //添加人uid
	CreatedName string `json:"created_name" db:"created_name" rule:"none"` //添加人name
	UpdatedAt   string `json:"updated_at" db:"updated_at" rule:"none"`     //更新时间
	UpdatedUID  string `json:"updated_uid" db:"updated_uid" rule:"none"`   //更新人uid
	UpdatedName string `json:"updated_name" db:"updated_name" rule:"none"` //更新人name
}

type MemberAssocLoginLogData struct {
	S int                   `json:"s"`
	D []MemberAssocLoginLog `json:"d"`
	T int64                 `json:"t"`
}

type MemberAssocLoginLog struct {
	Username string `json:"username"`
	IP       int64  `json:"ip"`
	IPS      string `json:"ips"`
	Device   string `json:"device"`
	DeviceNo string `json:"device_no"`
	Date     uint32 `json:"date"`
	Serial   string `json:"serial"`
	Agency   bool   `json:"agency"`
	Parents  string `json:"parents"`
	Tags     string `json:"tags"`
}

// 数据库 游戏字段
type GameLists struct {
	ID         string `db:"id" json:"id"`
	PlatformId string `db:"platform_id" json:"platform_id"`
	Name       string `db:"name" json:"name"`
	EnName     string `db:"en_name" json:"en_name"`
	ClientType string `db:"client_type" json:"client_type"`
	GameType   int64  `db:"game_type" json:"game_type"`
	GameId     string `db:"game_id" json:"game_id"`
	ImgPhone   string `db:"img_phone" json:"img_phone"`
	ImgPc      string `db:"img_pc" json:"img_pc"`
	ImgCover   string `db:"img_cover" json:"img_cover"`
	OnLine     int64  `db:"online" json:"online"`
	IsHot      int    `db:"is_hot" json:"is_hot"`
	IsNew      int    `db:"is_new" json:"is_new"`
	IsFs       int    `db:"is_fs" json:"is_fs"`
	Sorting    int64  `db:"sorting" json:"sorting"`
	CreatedAt  int64  `db:"created_at" json:"created_at"`
}

// 游戏列表返回数据结构
type GamePageList struct {
	D []GameLists `json:"d"`
	T int64       `json:"t"`
	S uint        `json:"s"`
}

type showGameJson struct {
	ID         string `db:"id" json:"id"`
	PlatformID string `db:"platform_id" json:"platform_id"`
	EnName     string `db:"en_name" json:"en_name"`
	ClientType string `db:"client_type" json:"client_type"`
	GameType   string `db:"game_type" json:"game_type"`
	GameID     string `db:"game_id" json:"game_id"`
	ImgPhone   string `db:"img_phone" json:"img_phone"`
	ImgPc      string `db:"img_pc" json:"img_pc"`
	IsHot      int    `db:"is_hot" json:"is_hot"`
	IsNew      int    `db:"is_new" json:"is_new"`
	Name       string `db:"name" json:"name"`
	ImgCover   string `db:"img_cover" json:"img_cover"`
	Sort       int    `db:"sorting" json:"sorting"`
	VnAlias    string `db:"vn_alias" json:"vn_alias"`
}

type Priv struct {
	ID        int64  `db:"id" json:"id" redis:"id"`                      //
	Name      string `db:"name" json:"name" redis:"name"`                //权限名字
	Module    string `db:"module" json:"module" redis:"module"`          //模块
	Sortlevel string `db:"sortlevel" json:"sortlevel" redis:"sortlevel"` //
	State     int    `db:"state" json:"state" redis:"state"`             //0:关闭1:开启
	Pid       int64  `db:"pid" json:"pid" redis:"pid"`                   //父级ID
}

type PrivTree struct {
	*Priv
	Parent *PrivTree `json:"parent"`
}

// 后台用户登录记录
type adminLoginLogBase struct {
	UID       string `msg:"uid" json:"uid"`
	Name      string `msg:"name" json:"name"`
	IP        string `msg:"ip" json:"ip"`
	Device    string `msg:"device" json:"device"`
	Flag      int    `msg:"flag" json:"flag"` // 1 登录 2 登出
	CreatedAt uint32 `msg:"created_at" json:"created_at"`
	Prefix    string `msg:"prefix" json:"prefix"`
}

type adminLoginLog struct {
	Id string `msg:"_id" json:"id"`
	adminLoginLogBase
}

// 后台用户登录记录
type AdminLoginLogData struct {
	D []adminLoginLog `json:"d"`
	T int64           `json:"t"`
	S int             `json:"s"`
}

// 系统日志
type systemLogBase struct {
	UID       string `msg:"uid" json:"uid"`
	Name      string `msg:"name" json:"name"`
	IP        string `msg:"ip" json:"ip"`
	Title     string `msg:"title" json:"title"`
	Content   string `msg:"content" json:"content"`
	CreatedAt uint32 `msg:"created_at" json:"created_at"`
	Prefix    string `msg:"prefix" json:"prefix"`
}

type systemLog struct {
	Id string `msg:"_id" json:"id"`
	systemLogBase
}

// 系统日志 分页展示数据
type SystemLogData struct {
	D []systemLog `json:"d"`
	T int64       `json:"t"`
	S int         `json:"s"`
}

type MemberRebate struct {
	UID       string `db:"uid" json:"uid"`
	ZR        string `db:"zr" json:"zr"` //真人返水
	QP        string `db:"qp" json:"qp"` //棋牌返水
	TY        string `db:"ty" json:"ty"` //体育返水
	DJ        string `db:"dj" json:"dj"` //电竞返水
	DZ        string `db:"dz" json:"dz"` //电游返水
	CreatedAt uint32 `db:"created_at" json:"created_at"`
	ParentUID string `db:"parent_uid" json:"parent_uid"`
	Prefix    string `db:"prefix" json:"prefix"`
}

type MemberMaxRebate struct {
	ZR sql.NullFloat64 `db:"zr" json:"zr"` //真人返水
	QP sql.NullFloat64 `db:"qp" json:"qp"` //棋牌返水
	TY sql.NullFloat64 `db:"ty" json:"ty"` //体育返水
	DJ sql.NullFloat64 `db:"dj" json:"dj"` //电竞返水
	DZ sql.NullFloat64 `db:"dz" json:"dz"` //电游返水
}

type NoticeData struct {
	D []Notice `json:"d"`
	T int64    `json:"t"`
	S uint     `json:"s"`
}

// 系统公告
type Notice struct {
	ID          string `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`               // 标题
	Content     string `db:"content" json:"content"`           // 内容
	Redirect    int    `db:"redirect" json:"redirect"`         // 是否跳转：1是 2否
	RedirectUrl string `db:"redirect_url" json:"redirect_url"` // 跳转url
	State       int    `db:"state" json:"state"`               // 1停用 2启用
	CreatedAt   int64  `db:"created_at" json:"created_at"`     // 创建时间
	CreatedUid  string `db:"created_uid" json:"created_uid"`
	CreatedName string `db:"created_name" json:"created_name"`
	Prefix      string `db:"prefix" json:"prefix"`
}

// 帐变数据
type TransactionData struct {
	T   int64         `json:"t"`
	D   []Transaction `json:"d"`
	Agg string        `db:"agg" json:"agg"`
}

type Transaction struct {
	ID           string `db:"id" json:"id" form:"id"`
	BillNo       string `db:"bill_no" json:"bill_no" form:"bill_no"`
	Uid          string `db:"uid" json:"uid" form:"uid"`
	Username     string `db:"username" json:"username" form:"username"`
	CashType     int    `db:"cash_type" json:"cash_type" form:"cash_type"`
	Amount       string `db:"amount" json:"amount" form:"amount"`
	BeforeAmount string `db:"before_amount" json:"before_amount" form:"before_amount"`
	AfterAmount  string `db:"after_amount" json:"after_amount" form:"after_amount"`
	CreatedAt    uint64 `db:"created_at" json:"created_at" form:"created_at"`
	Remark       string `db:"remark" json:"remark" form:"remark"`
}

// 场馆转账数据
type TransferData struct {
	T   int64      `json:"t"`
	D   []Transfer `json:"d"`
	Agg string     `db:"agg" json:"agg"`
}

//转账记录
type Transfer struct {
	ID           string `json:"id" db:"id"`
	UID          string `json:"uid" db:"uid"`
	BillNo       string `json:"bill_no" db:"bill_no"`
	PlatformId   string `json:"platform_id" db:"platform_id"`
	Username     string `json:"username" db:"username"`
	TransferType int    `json:"transfer_type" db:"transfer_type"`
	Amount       string `json:"amount" db:"amount"`
	BeforeAmount string `json:"before_amount" db:"before_amount"`
	AfterAmount  string `json:"after_amount" db:"after_amount"`
	CreatedAt    uint64 `json:"created_at" db:"created_at"`
	State        int    `json:"state" db:"state"`
	Automatic    int    `json:"automatic" db:"automatic"`
	ConfirmName  string `json:"confirm_name" db:"confirm_name"`
}

// 游戏记录数据
type GameRecordData struct {
	T   int64             `json:"t"`
	D   []GameRecord      `json:"d"`
	Agg map[string]string `json:"agg"`
}

//游戏投注记录结构
type GameRecord struct {
	ID             string  `db:"id" json:"id" form:"id"`
	RowId          string  `db:"row_id" json:"row_id" form:"row_id"`
	BillNo         string  `db:"bill_no" json:"bill_no" form:"bill_no"`
	ApiType        int     `db:"api_type" json:"api_type" form:"api_type"`
	ApiTypes       string  `json:"api_types"`
	PlayerName     string  `db:"player_name" json:"player_name" form:"player_name"`
	Name           string  `db:"name" json:"name" form:"name"`
	Uid            string  `db:"uid" json:"uid" form:"uid"`
	NetAmount      float64 `db:"net_amount" json:"net_amount" form:"net_amount"`
	BetTime        int64   `db:"bet_time" json:"bet_time" form:"bet_time"`
	StartTime      int64   `db:"start_time" json:"start_time" form:"start_time"`
	Resettle       uint8   `db:"resettle" json:"resettle" form:"resettle"`
	Presettle      uint8   `db:"presettle" json:"presettle" form:"presettle"`
	GameType       string  `db:"game_type" json:"game_type" form:"game_type"`
	BetAmount      float64 `db:"bet_amount" json:"bet_amount" form:"bet_amount"`
	ValidBetAmount float64 `db:"valid_bet_amount" json:"valid_bet_amount" form:"valid_bet_amount"`
	Flag           int     `db:"flag" json:"flag" form:"flag"`
	PlayType       string  `db:"play_type" json:"play_type" form:"play_type"`
	CopyFlag       int     `db:"copy_flag" json:"copy_flag" form:"copy_flag"`
	FilePath       string  `db:"file_path" json:"file_path" form:"file_path"`
	Prefix         string  `db:"prefix" json:"prefix" form:"prefix"`
	Result         string  `db:"result" json:"result" form:"result"`
	CreatedAt      uint64  `db:"created_at" json:"created_at" form:"created_at"`
	UpdatedAt      uint64  `db:"updated_at" json:"updated_at" form:"updated_at"`
	ApiName        string  `db:"api_name" json:"api_name" form:"api_name"`
	ApiBillNo      string  `db:"api_bill_no" json:"api_bill_no" form:"api_bill_no"`
	MainBillNo     string  `db:"main_bill_no" json:"main_bill_no" form:"main_bill_no"`
	IsUse          int     `db:"is_use" json:"is_use" form:"is_use"`
	FlowQuota      int64   `db:"flow_quota" json:"flow_quota" form:"flow_quota"`
	GameName       string  `db:"game_name" json:"game_name" form:"game_name"`
	HandicapType   string  `db:"handicap_type" json:"handicap_type" form:"handicap_type"`
	Handicap       string  `db:"handicap" json:"handicap" form:"handicap"`
	Odds           float64 `db:"odds" json:"odds" form:"odds"`
	BallType       int     `db:"ball_type" json:"ball_type" form:"ball_type"`
	SettleTime     int64   `db:"settle_time" json:"settle_time" form:"settle_time"`
	ApiBetTime     uint64  `db:"api_bet_time" json:"api_bet_time" form:"api_bet_time"`
	ApiSettleTime  uint64  `db:"api_settle_time" json:"api_settle_time" form:"api_settle_time"`
	IsRisk         int     `db:"-" json:"is_risk"`
	TopUid         string  `db:"top_uid" json:"top_uid,omitempty"`         //总代uid
	TopName        string  `db:"top_name" json:"top_name,omitempty"`       //总代代理
	ParentUid      string  `db:"parent_uid" json:"parent_uid,omitempty"`   //上级uid
	ParentName     string  `db:"parent_name" json:"parent_name,omitempty"` //上级代理
}

type Commissions struct {
	Id               string  `json:"id" db:"id"`
	Uid              string  `json:"uid" db:"uid"`
	Username         string  `json:"username" db:"username"`
	CreatedAt        int64   `json:"created_at" db:"created_at"`
	CommissionMonth  int64   `json:"commission_month" db:"commission_month"`
	TeamNum          int     `json:"team_num" db:"team_num"`
	ActiveNum        int     `json:"active_num" db:"active_num"`
	DepositAmount    float64 `json:"deposit_amount" db:"deposit_amount"`
	WithdrawAmount   float64 `json:"withdraw_amount" db:"withdraw_amount"`
	WinAmount        float64 `json:"win_amount" db:"win_amount"`
	PlatformAmount   float64 `json:"platform_amount" db:"platform_amount"`
	RebateAmount     float64 `json:"rebate_amount" db:"rebate_amount"`
	DividendAmount   float64 `json:"dividend_amount" db:"dividend_amount"`
	AdjustAmount     float64 `json:"adjust_amount" db:"adjust_amount"`
	NetWin           float64 `json:"net_win" db:"net_win"`
	BalanceAmount    float64 `json:"balance_amount" db:"balance_amount"`
	AdjustCommission float64 `json:"adjust_commission" db:"adjust_commission"`
	AdjustWin        float64 `json:"adjust_win" db:"adjust_win"`
	Amount           float64 `json:"amount" db:"amount"`
	Remark           string  `json:"remark" db:"remark"`
	Note             string  `json:"note" db:"note"`
	State            int     `json:"state" db:"state"`
	HandOutAt        int64   `json:"hand_out_at" db:"hand_out_at"`
	HandOutUid       string  `json:"hand_out_uid" db:"hand_out_uid"`
	HandOutName      string  `json:"hand_out_name" db:"hand_out_name"`
	DividendAgAmount float64 `json:"dividend_ag_amount" db:"dividend_ag_amount"`
	LastMonthAmount  float64 `json:"last_month_amount" db:"last_month_amount"`
	Prefix           string  `json:"prefix" db:"prefix"`
	PlanId           string  `json:"plan_id" db:"plan_id"`
	PlanName         string  `json:"plan_name" db:"plan_name"`
}

type CommissionTransaction struct {
	Id           string `json:"id" db:"id"`
	BillNo       string `json:"bill_no" db:"bill_no"`
	Uid          string `json:"uid" db:"uid"`
	Username     string `json:"username" db:"username"`
	CashType     int    `json:"cash_type" db:"cash_type"`
	Amount       string `json:"amount" db:"amount"`
	BeforeAmount string `json:"before_amount" db:"before_amount"`
	AfterAmount  string `json:"after_amount" db:"after_amount"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	Prefix       string `json:"prefix" db:"prefix"`
}

type MembersTree struct {
	Ancestor   string `db:"ancestor"`
	Descendant string `db:"descendant"`
	Lvl        int    `db:"lvl"`
}

type CommssionConf struct {
	ID     string `json:"id" db:"id"`
	UID    string `json:"uid" db:"uid"`
	PlanID string `json:"plan_id" db:"plan_id"`
}

// CommissionPlan 返佣方案具体比例
type CommissionPlan struct {
	ID              string `db:"id" json:"id"`
	Name            string `db:"name" json:"name"`                         //方案名称
	CommissionMonth int64  `db:"commission_month" json:"commission_month"` // 生效月份
	CreatedAt       int64  `db:"created_at" json:"created_at"`
	UpdatedUID      string `db:"updated_uid" json:"updated_uid"`
	UpdatedName     string `db:"updated_name" json:"updated_name"`
	UpdatedAt       int64  `db:"updated_at" json:"updated_at"`
	Prefix          string `db:"prefix" json:"prefix"`
}

// CommissionDetail 返佣方案具体返水
type CommissionDetail struct {
	ID     string  `db:"id" json:"id"`
	PlanID string  `db:"plan_id" json:"plan_id"` // 所属方案
	WinMax float64 `db:"win_max" json:"win_max"` //净输赢最大值
	WinMin float64 `db:"win_min" json:"win_min"` //净输赢最小值
	Rate   float64 `db:"rate" json:"rate"`       //返佣比例
	Prefix string  `db:"prefix" json:"prefix"`
}

type CommPlanPageData struct {
	T       int64                         `json:"t"`
	D       []CommissionPlan              `json:"d"`
	S       uint                          `json:"s"`
	Details map[string][]CommissionDetail `json:"details"`
}
