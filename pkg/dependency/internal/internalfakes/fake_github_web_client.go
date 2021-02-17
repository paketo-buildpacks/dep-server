// Code generated by counterfeiter. DO NOT EDIT.
package internalfakes

import (
	"sync"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

type FakeGithubWebClient struct {
	DownloadStub        func(string, string, ...internal.RequestOption) error
	downloadMutex       sync.RWMutex
	downloadArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 []internal.RequestOption
	}
	downloadReturns struct {
		result1 error
	}
	downloadReturnsOnCall map[int]struct {
		result1 error
	}
	GetStub        func(string, ...internal.RequestOption) ([]byte, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 string
		arg2 []internal.RequestOption
	}
	getReturns struct {
		result1 []byte
		result2 error
	}
	getReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	PostStub        func(string, []byte, ...internal.RequestOption) ([]byte, error)
	postMutex       sync.RWMutex
	postArgsForCall []struct {
		arg1 string
		arg2 []byte
		arg3 []internal.RequestOption
	}
	postReturns struct {
		result1 []byte
		result2 error
	}
	postReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGithubWebClient) Download(arg1 string, arg2 string, arg3 ...internal.RequestOption) error {
	fake.downloadMutex.Lock()
	ret, specificReturn := fake.downloadReturnsOnCall[len(fake.downloadArgsForCall)]
	fake.downloadArgsForCall = append(fake.downloadArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 []internal.RequestOption
	}{arg1, arg2, arg3})
	stub := fake.DownloadStub
	fakeReturns := fake.downloadReturns
	fake.recordInvocation("Download", []interface{}{arg1, arg2, arg3})
	fake.downloadMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGithubWebClient) DownloadCallCount() int {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	return len(fake.downloadArgsForCall)
}

func (fake *FakeGithubWebClient) DownloadCalls(stub func(string, string, ...internal.RequestOption) error) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = stub
}

func (fake *FakeGithubWebClient) DownloadArgsForCall(i int) (string, string, []internal.RequestOption) {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	argsForCall := fake.downloadArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeGithubWebClient) DownloadReturns(result1 error) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = nil
	fake.downloadReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeGithubWebClient) DownloadReturnsOnCall(i int, result1 error) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = nil
	if fake.downloadReturnsOnCall == nil {
		fake.downloadReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.downloadReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeGithubWebClient) Get(arg1 string, arg2 ...internal.RequestOption) ([]byte, error) {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 string
		arg2 []internal.RequestOption
	}{arg1, arg2})
	stub := fake.GetStub
	fakeReturns := fake.getReturns
	fake.recordInvocation("Get", []interface{}{arg1, arg2})
	fake.getMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGithubWebClient) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeGithubWebClient) GetCalls(stub func(string, ...internal.RequestOption) ([]byte, error)) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *FakeGithubWebClient) GetArgsForCall(i int) (string, []internal.RequestOption) {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeGithubWebClient) GetReturns(result1 []byte, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeGithubWebClient) GetReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeGithubWebClient) Post(arg1 string, arg2 []byte, arg3 ...internal.RequestOption) ([]byte, error) {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.postMutex.Lock()
	ret, specificReturn := fake.postReturnsOnCall[len(fake.postArgsForCall)]
	fake.postArgsForCall = append(fake.postArgsForCall, struct {
		arg1 string
		arg2 []byte
		arg3 []internal.RequestOption
	}{arg1, arg2Copy, arg3})
	stub := fake.PostStub
	fakeReturns := fake.postReturns
	fake.recordInvocation("Post", []interface{}{arg1, arg2Copy, arg3})
	fake.postMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGithubWebClient) PostCallCount() int {
	fake.postMutex.RLock()
	defer fake.postMutex.RUnlock()
	return len(fake.postArgsForCall)
}

func (fake *FakeGithubWebClient) PostCalls(stub func(string, []byte, ...internal.RequestOption) ([]byte, error)) {
	fake.postMutex.Lock()
	defer fake.postMutex.Unlock()
	fake.PostStub = stub
}

func (fake *FakeGithubWebClient) PostArgsForCall(i int) (string, []byte, []internal.RequestOption) {
	fake.postMutex.RLock()
	defer fake.postMutex.RUnlock()
	argsForCall := fake.postArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeGithubWebClient) PostReturns(result1 []byte, result2 error) {
	fake.postMutex.Lock()
	defer fake.postMutex.Unlock()
	fake.PostStub = nil
	fake.postReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeGithubWebClient) PostReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.postMutex.Lock()
	defer fake.postMutex.Unlock()
	fake.PostStub = nil
	if fake.postReturnsOnCall == nil {
		fake.postReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.postReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeGithubWebClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.postMutex.RLock()
	defer fake.postMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGithubWebClient) recordInvocation(key string, args []interface{}) {
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

var _ internal.GithubWebClient = new(FakeGithubWebClient)