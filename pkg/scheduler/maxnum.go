package scheduler

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

/*
	自定义调度插件：自定义最大POD数量的调度插件
*/

const TestSchedulingName = "test-pod-maxnum-scheduler"

// TestPodNumScheduling 调度器插件对象
type TestPodNumScheduling struct {
	fact informers.SharedInformerFactory
	args *Args
}

// Args 配置文件参数
type Args struct {
	MaxPods int `json:"maxPods,omitempty"`
}

func (s *TestPodNumScheduling) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (s *TestPodNumScheduling) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

// 通过label过滤不能调度的节点
const (
	SchedulingLabelKeyState   = "scheduling"
	SchedulingLabelValueState = "true"
)

// Filter 过滤方法 (过滤node条件)
func (s *TestPodNumScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {

	for k, v := range nodeInfo.Node().Labels {
		if k == SchedulingLabelKeyState && v != SchedulingLabelValueState {
			klog.V(3).Info("This node is unschedulable")
			return framework.NewStatus(framework.Unschedulable, "This node is unschedulable")
		}
	}
	return framework.NewStatus(framework.Success)

}

// PreFilter 前置过滤方法 (过滤pod条件)
func (s *TestPodNumScheduling) PreFilter(_ context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.V(3).Infof("prefilter step start [Pod] %v\n", p.Name)
	// informer list pod
	// 通过informer函数列出指定命名空间下的所有Pod
	podList, err := s.fact.Core().V1().Pods().Lister().Pods(p.Namespace).List(labels.Everything())
	if err != nil {
		klog.Errorf("pod informer list error: \n", err)
		return framework.NewStatus(framework.Error)
	}

	// 过滤
	// 若配置的MaxPods参数大于0且Pods数量大于该值
	if s.args.MaxPods > 0 && len(podList) > s.args.MaxPods {
		klog.V(3).Infof("The number of [Pod] %v exceeds the schedulable number and cannot be scheduled\n", p.Name)
		return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("POD数量超过，不能调度，最多只能调度%d", s.args.MaxPods))
	}
	klog.V(3).Infof("[Pod] %v successfully scheduled\n", p.Name)
	return framework.NewStatus(framework.Success)
}

func (s *TestPodNumScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return s
}

func (*TestPodNumScheduling) Name() string {
	return TestSchedulingName
}

// 检查是否实现接口对象
var _ framework.PreFilterPlugin = &TestPodNumScheduling{}
var _ framework.FilterPlugin = &TestPodNumScheduling{}

func NewTestPodNumScheduling(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	// 注入配置文件参数
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	return &TestPodNumScheduling{
		fact: f.SharedInformerFactory(),
		args: args,
	}, nil
}
