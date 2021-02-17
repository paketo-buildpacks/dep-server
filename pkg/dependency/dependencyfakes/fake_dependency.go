// Code generated by counterfeiter. DO NOT EDIT.
package dependencyfakes

import (
	"sync"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
)

type FakeDependency struct {
	GetAllVersionRefsStub        func() ([]string, error)
	getAllVersionRefsMutex       sync.RWMutex
	getAllVersionRefsArgsForCall []struct {
	}
	getAllVersionRefsReturns struct {
		result1 []string
		result2 error
	}
	getAllVersionRefsReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	GetDependencyVersionStub        func(string) (dependency.DepVersion, error)
	getDependencyVersionMutex       sync.RWMutex
	getDependencyVersionArgsForCall []struct {
		arg1 string
	}
	getDependencyVersionReturns struct {
		result1 dependency.DepVersion
		result2 error
	}
	getDependencyVersionReturnsOnCall map[int]struct {
		result1 dependency.DepVersion
		result2 error
	}
	GetReleaseDateStub        func(string) (time.Time, error)
	getReleaseDateMutex       sync.RWMutex
	getReleaseDateArgsForCall []struct {
		arg1 string
	}
	getReleaseDateReturns struct {
		result1 time.Time
		result2 error
	}
	getReleaseDateReturnsOnCall map[int]struct {
		result1 time.Time
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDependency) GetAllVersionRefs() ([]string, error) {
	fake.getAllVersionRefsMutex.Lock()
	ret, specificReturn := fake.getAllVersionRefsReturnsOnCall[len(fake.getAllVersionRefsArgsForCall)]
	fake.getAllVersionRefsArgsForCall = append(fake.getAllVersionRefsArgsForCall, struct {
	}{})
	stub := fake.GetAllVersionRefsStub
	fakeReturns := fake.getAllVersionRefsReturns
	fake.recordInvocation("GetAllVersionRefs", []interface{}{})
	fake.getAllVersionRefsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDependency) GetAllVersionRefsCallCount() int {
	fake.getAllVersionRefsMutex.RLock()
	defer fake.getAllVersionRefsMutex.RUnlock()
	return len(fake.getAllVersionRefsArgsForCall)
}

func (fake *FakeDependency) GetAllVersionRefsCalls(stub func() ([]string, error)) {
	fake.getAllVersionRefsMutex.Lock()
	defer fake.getAllVersionRefsMutex.Unlock()
	fake.GetAllVersionRefsStub = stub
}

func (fake *FakeDependency) GetAllVersionRefsReturns(result1 []string, result2 error) {
	fake.getAllVersionRefsMutex.Lock()
	defer fake.getAllVersionRefsMutex.Unlock()
	fake.GetAllVersionRefsStub = nil
	fake.getAllVersionRefsReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) GetAllVersionRefsReturnsOnCall(i int, result1 []string, result2 error) {
	fake.getAllVersionRefsMutex.Lock()
	defer fake.getAllVersionRefsMutex.Unlock()
	fake.GetAllVersionRefsStub = nil
	if fake.getAllVersionRefsReturnsOnCall == nil {
		fake.getAllVersionRefsReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.getAllVersionRefsReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) GetDependencyVersion(arg1 string) (dependency.DepVersion, error) {
	fake.getDependencyVersionMutex.Lock()
	ret, specificReturn := fake.getDependencyVersionReturnsOnCall[len(fake.getDependencyVersionArgsForCall)]
	fake.getDependencyVersionArgsForCall = append(fake.getDependencyVersionArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetDependencyVersionStub
	fakeReturns := fake.getDependencyVersionReturns
	fake.recordInvocation("GetDependencyVersion", []interface{}{arg1})
	fake.getDependencyVersionMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDependency) GetDependencyVersionCallCount() int {
	fake.getDependencyVersionMutex.RLock()
	defer fake.getDependencyVersionMutex.RUnlock()
	return len(fake.getDependencyVersionArgsForCall)
}

func (fake *FakeDependency) GetDependencyVersionCalls(stub func(string) (dependency.DepVersion, error)) {
	fake.getDependencyVersionMutex.Lock()
	defer fake.getDependencyVersionMutex.Unlock()
	fake.GetDependencyVersionStub = stub
}

func (fake *FakeDependency) GetDependencyVersionArgsForCall(i int) string {
	fake.getDependencyVersionMutex.RLock()
	defer fake.getDependencyVersionMutex.RUnlock()
	argsForCall := fake.getDependencyVersionArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDependency) GetDependencyVersionReturns(result1 dependency.DepVersion, result2 error) {
	fake.getDependencyVersionMutex.Lock()
	defer fake.getDependencyVersionMutex.Unlock()
	fake.GetDependencyVersionStub = nil
	fake.getDependencyVersionReturns = struct {
		result1 dependency.DepVersion
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) GetDependencyVersionReturnsOnCall(i int, result1 dependency.DepVersion, result2 error) {
	fake.getDependencyVersionMutex.Lock()
	defer fake.getDependencyVersionMutex.Unlock()
	fake.GetDependencyVersionStub = nil
	if fake.getDependencyVersionReturnsOnCall == nil {
		fake.getDependencyVersionReturnsOnCall = make(map[int]struct {
			result1 dependency.DepVersion
			result2 error
		})
	}
	fake.getDependencyVersionReturnsOnCall[i] = struct {
		result1 dependency.DepVersion
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) GetReleaseDate(arg1 string) (time.Time, error) {
	fake.getReleaseDateMutex.Lock()
	ret, specificReturn := fake.getReleaseDateReturnsOnCall[len(fake.getReleaseDateArgsForCall)]
	fake.getReleaseDateArgsForCall = append(fake.getReleaseDateArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetReleaseDateStub
	fakeReturns := fake.getReleaseDateReturns
	fake.recordInvocation("GetReleaseDate", []interface{}{arg1})
	fake.getReleaseDateMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDependency) GetReleaseDateCallCount() int {
	fake.getReleaseDateMutex.RLock()
	defer fake.getReleaseDateMutex.RUnlock()
	return len(fake.getReleaseDateArgsForCall)
}

func (fake *FakeDependency) GetReleaseDateCalls(stub func(string) (time.Time, error)) {
	fake.getReleaseDateMutex.Lock()
	defer fake.getReleaseDateMutex.Unlock()
	fake.GetReleaseDateStub = stub
}

func (fake *FakeDependency) GetReleaseDateArgsForCall(i int) string {
	fake.getReleaseDateMutex.RLock()
	defer fake.getReleaseDateMutex.RUnlock()
	argsForCall := fake.getReleaseDateArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeDependency) GetReleaseDateReturns(result1 time.Time, result2 error) {
	fake.getReleaseDateMutex.Lock()
	defer fake.getReleaseDateMutex.Unlock()
	fake.GetReleaseDateStub = nil
	fake.getReleaseDateReturns = struct {
		result1 time.Time
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) GetReleaseDateReturnsOnCall(i int, result1 time.Time, result2 error) {
	fake.getReleaseDateMutex.Lock()
	defer fake.getReleaseDateMutex.Unlock()
	fake.GetReleaseDateStub = nil
	if fake.getReleaseDateReturnsOnCall == nil {
		fake.getReleaseDateReturnsOnCall = make(map[int]struct {
			result1 time.Time
			result2 error
		})
	}
	fake.getReleaseDateReturnsOnCall[i] = struct {
		result1 time.Time
		result2 error
	}{result1, result2}
}

func (fake *FakeDependency) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getAllVersionRefsMutex.RLock()
	defer fake.getAllVersionRefsMutex.RUnlock()
	fake.getDependencyVersionMutex.RLock()
	defer fake.getDependencyVersionMutex.RUnlock()
	fake.getReleaseDateMutex.RLock()
	defer fake.getReleaseDateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDependency) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ dependency.Dependency = new(FakeDependency)