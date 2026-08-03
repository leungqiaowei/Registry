package apis

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/registry/data"

	//"hit.edu/framework/pkg/registry/data"
	"time"
)

// 定义框架基础资源
// 节点资源信息
// TODO: 接口版本

// TODO: 独立配置
type Time struct {
	time.Time
}

// TODO: 独立配置
type Event struct {
	//TODO: 定义Event
	//TODO: ObjectReference设计
}

// Node
type Node struct {
	//
	runtime.TypeMeta

	//
	meta.ObjectMeta

	// 定义Node行为
	Spec NodeSpec

	// 定义Node的当前状态
	Status NodeStatus
}

type NodeSpec struct {
	// 调度时需要使用，将任务调度到该节点
	NodeName string

	// 节点的HostName
	HostName string

	// 不能被调度的节点
	Unschedulable bool
	// TODO: 节点Label
}

type NodeList struct {
	runtime.TypeMeta

	// TODO: List Options

	Items []Node
}

// 计算、网络、存储等定量资源
// 资源名称 => 定量资源描述
type ResourceList map[string]Quantity

type NodeStatus struct {
	// 节点上的所有物理资源
	Capacity ResourceList

	// 节点上当前可以分配的资源
	Allocatable ResourceList

	// 这里只表示连接关系，对硬件的调用放到能力中
	// TODO: 节点上的硬件资源

	// TODO：节点上部署的任务

	// 节点上的Docker镜像
	Images []ContainerImage

	// 节点上的Wasm镜像
	Wasms []WasmImage

	// TODO：节点上的依赖情况

	// 节点的物理地址，可以有多个
	Addresses NodeAddress

	// 节点系统信息
	NodeInfo NodeSystemInfo
}

// From K8s
type NodeSystemInfo struct {
	// MachineID reported by the node. For unique machine identification
	// in the cluster this field is preferred. Learn more from man(5)
	MachineID string
	// SystemUUID reported by the node. For unique machine identification
	// MachineID is preferred. This field is specific to Red Hat hosts
	SystemUUID string
	// Boot ID reported by the node.
	BootID string
	// Kernel Version reported by the node.
	KernelVersion string
	// OS Image reported by the node.
	OSImage string
	// ContainerRuntime Version reported by the node.
	ContainerRuntimeVersion string
	// NodeLet
	NodeletVersion string
	// The Operating System reported by the node
	OperatingSystem string
	// The Architecture reported by the node
	Architecture string
}

// NodeAddress represents node's address
type NodeAddress struct {
	Type    string
	Address string
}

// From K8s
type ContainerImage struct {
	// Names by which this image is known.
	Names []string
	// The size of the image in bytes.
	SizeBytes int64
}

// Wasm镜像
type WasmImage struct {
	// Names by which this image is known.
	Names []string
	// The size of the image in bytes.
	SizeBytes int64
}

// 设备资源信息
// Device
// Device具有Ability

// 能力资源信息
// Ability
// TODO: 待增加

// 工作流相关资源，表示一次部署执行的内容
// 框架中工作流包含四级,Workflow => Task => Group => Action
// 其中Group和Action具体部署在节点上
// Task和Workflow表示逻辑结构,Group,Action表示具体的执行关系

// 对工作流的描述
type Description struct {
	// TODO: Label单独字段
	Label []string
	// 用户对工作流行为的描述
	// +Optional
	Docs string
}

// 工作流节点执行需要满足以下条件
// 1. 节点依赖，所有前序节点都需要执行完成
// 2. 数据依赖，所需的数据都下载到对应的节点上
// 3. 资源依赖，所需的资源都已经满足
// 4. 程序依赖，程序依赖已经安装完成

// TODO: 增加Label及相关选择器

// ------- Workflow

// 工作流相关生命周期
type Phase string

const (
	// 任务相关状态
	Pending   Phase = "Pending"
	Running   Phase = "Running"
	Successed Phase = "Succeeded"
	Failed    Phase = "Failed"
	Unknown   Phase = "Unknown"
	// 迁移相关状态
	Migrating Phase = "Migrating"
	Migrated  Phase = "Migrated"
)

// 定义流程类型，用于表示有条件的DAG
// 支持顺序、分支和循环（有限展开)
type ProcessType string

const (
	// 普通类型的节点，默认为该节点
	Norm ProcessType = "Normal"
	// 需要判断执行条件的Action
	// 用于分支类型的节点
	Cond ProcessType = "Condition"
	// 带有循环生成器的Action
	// 检查Condition是否满足所需条件
	// 如果不满足条件，或者未达到最大循环次数，则继续生成Action
	// TODO: 生成的ActionID需要以Loop+{Count}为后缀
	// TODO: 循环计数器Count
	Loop ProcessType = "Loop"
	//...待后续拓展
)

// 流程条件

type ConditionValueType string

const (
	Constants ConditionValueType = "Constants"
	Dynamic   ConditionValueType = "Dynamic"
)

// TODO: 参考Inputs,重新定义
type ConditionValue struct {
	// Condition的变量有以下类型
	// 	Constants, Value则为常数，From中信息为空
	//  Dynamic, Value需要根据From中的信息从Results中获取
	//  	比如前序任务的执行状态
	//		或者循环计数器的数量
	// 		或者任务的执行结果
	// TODO: 使用DataType代替，更新相关文档
	Type DataType //ConditionValueType
	//
	Name string
	// 实际的值
	Value string
	// TODO: From
	// TODO: 动态类型的Value,数据来源,需要对应的Controller Watch相关变量
	// +Optional
	ValueType string
}

// 条件连接符，支持大小写
type JoinType string

const (
	and JoinType = "and"
	or  JoinType = "or"
	And JoinType = "and"
	Or  JoinType = "or"
)

// 符号判断
type SignalType string

const (
	Equal    SignalType = "=="
	NotEqual SignalType = "!="
)

// 流程执行条件
// LeftValue ==或!= RightValue
// 输出结果为Bool类型的值
// TODO: Value格式检查和调整，比如存在空格的情况
type ConditionFormula struct {
	LeftValue  ConditionValue
	RightValue ConditionValue
	// == 或 !=
	Signal SignalType
	// 在条件串中的期望结果
	// 类型包含 and 或者 or
	Join JoinType
	// TODO: 符号判断结果
}

// 不建议使用过于复杂的逻辑
// 只支持逻辑的串行连接
// 如果没有Condition,Conditions默认值为true
// Example: Condition[1] and/or Condition[2] and/or Condition[3] ......
type Conditions struct {
	Formulas []ConditionFormula
}

// 工作流相关Ref
// 表示一个工作流属于哪个ID
// 唯一标识
type IDRef struct {
	WorkflowID string
	TaskID     string
	GroupID    string
	ActionID   string
}

type Workflow struct {
	//
	runtime.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec WorkflowSpec

	//
	Status WorkflowStatus
}

type WorkflowSpec struct {
	// Workflow Name, 用户提交时提供的Name
	// FIXME: Name应当唯一，否则不允许提交
	Name string

	// 对工作流的描述
	Desc Description

	// 工作流下有哪些任务
	// 目前只支持Task=>Group=>Action
	// TODO: 更为宽松的Task关系定义
	// 目前与Action的定义方式类似
	Tasks []Task
	// TODO: Tasks的拓扑关系
}

type WorkflowStatus struct {
	// Workflow ID，系统为根据Desc中的Name为Workflow分配的唯一标识, 用户填写时应该为空
	WorkflowID string

	//
	Phase Phase

	//
	TaskStatus []TaskStatus

	// 执行时间
	StartAt Time
	// 结束时间
	FinishAt Time
	// 最新获取状态的时间
	LastTime Time
}

// --------- Task
type Task struct {
	//
	runtime.TypeMeta

	//
	meta.ObjectMeta

	Spec   TaskSpec
	Status TaskStatus
}

type TaskTemplate struct {
	Spec TaskSpec
}

type TaskSpec struct {
	// Task Name
	Name string

	// Parents Name
	Parents []string

	// Task描述
	Desc Description

	// Task类型
	Type ProcessType

	// Task条件
	// +Optional
	Conditions Conditions

	// 存储当前Task中的所有Group
	// Group之间没有严格的依赖限制
	// 可以有多个Group作为Group的入口，支持多个图形结构
	// 支持并发执行多组Group
	// TODO: 增加Level支持
	Groups []Group
}

type TaskStatus struct {
	//
	TaskID string

	//
	Belongs IDRef

	//
	Phase Phase

	//
	GroupStatus []GroupStatus

	// TODO: Events定义

	// 执行时间
	StartAt Time
	// 结束时间
	FinishAt Time
	// 最新获取状态的时间
	LastTime Time
}

// ---------- Group
type Group struct {
	//
	runtime.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec GroupSpec

	//
	Status GroupStatus
}

type GroupTemplate struct {
	Spec GroupSpec
}

type GroupSpec struct {
	// Group Name
	Name string

	// Parents Name, 一个Group可以有多个Parents
	Parents []string

	// Group描述
	Desc Description

	// Group类型
	Type ProcessType

	// Group条件
	// +Optional
	Conditions Conditions

	// Group所需资源需要先遍历自己的Action
	//   Ref类型的指针需要根据Action中的资源需求计算
	//   TODO: 对于Condition类的节点，使用资源的预估
	//   TODO: 资源需求

	// 存储当前Group所属的所有Action
	// Group中可以只有一个Action, 简化实现逻辑
	Actions []Action
	// TODO: 生成时是否可以直接分析依赖关系？
	// TODO: Action的依赖关系描述, 需要先遍历Actions构建DAG图
	// Group中的Action需要有较严格的依赖顺序，可以支持分支,条件,循环
	// 只能有一个Action作为入口Action,图形结构
}

type GroupStatus struct {
	//
	GroupID string

	//
	Belongs IDRef

	//
	Phase Phase

	// TODO: 所属Actions的状态
	ActionStatus []ActionStatus

	// TODO: 整体资源使用情况

	// TODO: Events定义

	// 执行时间
	StartAt Time
	// 结束时间
	FinishAt Time
	// 最新获取状态的时间
	LastTime Time
}

// ---------- Action

type Action struct {
	//
	runtime.TypeMeta

	//
	meta.ObjectMeta

	//
	Spec ActionSpec

	//
	Status ActionStatus
}

// 创建Action所需要的字段
type ActionTemplate struct {
	Spec ActionSpec
}

type ActionSpec struct {
	// Action Name
	Name string

	// Parents Name, 一个节点可以有多个Parents
	Parents []string

	// Action描述
	Desc Description

	// Action类型
	Type ProcessType

	// Action条件
	// +Optional
	Conditions Conditions

	// 需要的执行环境
	// 串行执行Runtime中的运行环境
	// 如果使用Pod方式部署的任务，建议只部署一个阻塞的Runtime
	// 可以用于执行命令，环境准备等操作
	Runtimes []Runtime

	// 其他配置/选项
	// 细粒度任务控制
	// TODO: RPC相关接口实现
	EnableFineGrainedControl bool
}

type RuntimeType string

const (
	ByDevice     RuntimeType = "device"
	ByNet        RuntimeType = "net"
	ByCommand    RuntimeType = "command"
	ByBinary     RuntimeType = "binary"
	ByDocker     RuntimeType = "docker"
	ByService    RuntimeType = "service"
	ByDeployment RuntimeType = "deployment"
	ByPod        RuntimeType = "pod"
)

// 环境变量
// TODO: 环境变量格式定义
type EnvVar struct {
	Name  string
	Value string
	// TODO: 动态获取相关字段
}

// TODO: 后续补充完整
type ResourceSpec struct{}
type ResourceStatus struct{}

type DeviceSpec struct{}
type DeviceStatus struct{}

type DataSpec struct {
	// 这个结构体加上文件的详细数据，便于统一查询，在文件服务器存储文件时按照这种格式来。
	// 文件名称
	FileName string `json:"file_name"`
	// 文件版本号
	Tag string `json:"tag"`
	// 对于文件类型的Data
	Type data.FileType `json:"type"`
	// 标识文件是否需要持久化存储
	IsPermanent bool `json:"is_permanent"`
	// 文件在文件服务器中的存储路径
	FilePath string `json:"file_path"`
	// 存储的时间
	StorageTime time.Time `json:"storage_time"`
	// 文件所有者（上传文件方的ip？如果能获取的话）
	Owner string `json:"owner"`
	// 文件大小
	Size int64 `json:"size"`
	// 文件哈希值（SHA文件校验，通过对文件内容进行哈希计算（SHA-256）得到的固定长度的字符串。哈希值可以用于验证文件的完整性，确保文件在传输或存储过程中没有被篡改）
	FileHash string `json:"file_hash"`
}
type DataStatus struct{}

// 文件的状态：待订阅和已定阅
// 在处理过程中，或者已经处理完成。
type SceneSpec struct{}
type SceneStatus struct{}

// Action所需执行环境
type Runtime struct {
	// 定义Action所需资源
	// 在Group调度完成后，Action需要被分配和保留，直到Action开始执行
	// 动态资源分配所需字段

	// 运行时,需要检查以下四类资源，只有资源都满足时，才可以继续执行
	// 需要的计算、网络、存储资源
	Resources []ResourceSpec

	// 需要的硬件资源
	Devices []DeviceSpec

	// 需要使用的数据
	Data []DataSpec

	// 需要使用的场景数据
	Scenes []SceneSpec

	// 运行时类型
	// 执行环境,有如下种类：
	//      Device, 端侧需要控制设备和能力
	//      Net, 需要发起网络请求
	//      Command, 需要执行脚本命令
	// 		Binary, 需要检查环境依赖，二进制所在位置
	// 		Docker, 需要下载镜像到对应节点，启动Docker并注入环境依赖
	// 		Service, 通过K8s部署Service到节点上
	// 		Deployment, 通过K8s部署Deployment到节点上
	// 		Pod, 通过K8s部署Pod到节点上
	// 其中端侧只能使用Device,Net,Command,Binary或者Docker的方式部署运行环境,使用Docker环境部署需要注意网络环境
	// 云边侧支持除Device外的其他方式,使用Docker部署需要注意网络环境
	// 通过Pod方式部署的Action,如果不启用细粒度任务控制，需要通过K8s相关接口获取运行状态
	// TODO: 如果使用Pod方式部署，需要记录Pod对应的ID,如果生命周期内Pod存在改动，该需要对应的更新
	// TODO: 实现一个Controller,专门监控对应的Pod的变化
	// TODO: 使用Pod方式部署，可以同时部署Action的多个副本
	Type RuntimeType

	// 标识运行时的名称
	Name string

	// 镜像
	//  对于Command,Net类型，Image应当为空
	// 	对于Binary或者Script，Image对应二进制上传的位置，部署时需要检查依赖环境
	// 	对于Docker,Image对应镜像的存储位置
	//  对于Service,Deployment和Pod,对应相关文件的存储位置，需要使用网络链接
	// 需要先检查本地是否有对应版本的镜像
	// +Optional
	Image string

	// TODO: 软件依赖如何表示

	// 执行参数
	// 程序的入口函数，一般不做更改
	Command []string

	// 如果程序要注入其他的运行参数，则放到这里
	// +Optional
	Args []string

	// 环境变量
	// +Optional
	EnvVar []EnvVar

	// 需要的数据
	// 输入数据
	// 	输入数据作为参数注入到命令参数中
	Inputs Input

	// 输出数据
	//  输出数据作为参数注入到命令参数中
	Outputs Output
}

//	 输入的数据有以下几类
//			常量类型的数据
//			从Results中获取数据
//			从本地获取的资源中获取数据
type DataType string

const (
	ConstData   DataType = "constants"
	ResultsData DataType = "results"
	LocalData   DataType = "local"
)

type Input struct {
	// 类型
	Type DataType

	// 名称
	Name string

	// 值
	//   对应常量类型，ValueType对应常量的类型，Value对应常量的值，类型与值应该对应，支持的类型
	// 		包括：整数、浮点数、布尔值
	//   对于Results类型，Value对应从哪个节点的Results中获取数据，ValueType对应从Results中获取的ResultType
	//      Results的访问格式对应
	// 			TODO: 正则表达式
	//			Action{ID}.Results.{Name}， 缺省访问本Group对应的Action
	//			Group{ID}.Action{ID}.Results.{Name}， 访问对应Group的对应Action
	// 			Task{ID}.Group{ID}.Action{ID}.Results.{Name}, 访问对应Task的对应Group的对应Action
	//          目前不支持跨Workflow获取数据
	//   对于Local类型，Value对应从本地资源或者数据节点获取的数据，ValueType对应从资源或者数据中获取的数据类型
	//   	Local的访问格式对应
	//          Resources.{Name}: 从本地资源中获取
	//			Devices.{Name}: 从本地设备里列表中获取
	//			Scenes.{Name}： 从本地场景中获取
	//			Data.{Name}： 从本地数据中获取
	//      Local类型的数据对其他节点不可见
	Value     string
	ValueType string
}

// TODO: 数据格式后续还需要调整
type Output struct {
	//
	Type DataType

	Name string

	// TODO:
	Value     string
	ValueType string
}

type ActionStatus struct {
	// 在开始执行时，确定Action所属的Workflow,Task,Group
	// ID组成格式为ActionName+GroupID
	ActionID string
	// Action所属标识
	Belongs IDRef
	// 生命周期
	Phase Phase
	// 当前资源使用情况
	Resources []ResourceStatus
	// 当前设备使用情况
	Devices []DeviceStatus
	// 当前数据使用情况
	Data []DataStatus
	// 当前场景更新情况
	Scenes []SceneStatus
	// 当前Runtime执行状态
	RuntimeStatus []RuntimeStatus
	// 任务执行结果
	Results []Result
	// TODO: Events定义

	// 执行时间
	StartAt Time
	// 结束时间
	FinishAt Time
	// 最新获取状态的时间
	LastTime Time
}

type RuntimeStatus struct {
	// 当前任务的执行情况
	// 任务在哪里执行,进程ID
	NodeName  string
	ProcessId string

	// 执行状态
	Phase Phase

	// 执行时间
	StartAt Time
	// 结束时间
	FinishAt Time
	// 最新获取状态的时间
	LastTime Time
}

// 任务的输出结果
// 单独作为一个资源，方便其他节点获取该资源
// 数据结果的类型
type Result struct {
	//Result唯一表示，由ActionID
	ResultID string
	// Result属于哪个Action/Group/Task/Workflow
	Belongs IDRef
	// Results类型，有如下的类型
	// 		存在内存中的结果,可以通过etcd模块同步
	// 		比较大的数据，需要提供资源的访问方式，资源同步模块需要该字段，由该模块实现资源的点对点同步
	Type string
	// TODO: Results存储位置，访问方式
	// Results可以有备份位置
	// 需要构建Ownership表，任务执行节点可以获取所需数据的位置
}

// 系统运行时相关信息，表示长时间部署执行的内容
// 云边侧模块使用相关字段
// TODO: Pod From K8s
type Pod struct{}

// TODO: Service From K8s
type Service struct{}

// TODO: Deployment From K8s
type Deployment struct{}

// TODO: VM From K8s
type VM struct{}

// 后续需要可以扩展
