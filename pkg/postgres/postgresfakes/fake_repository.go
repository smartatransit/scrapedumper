// Code generated by counterfeiter. DO NOT EDIT.
package postgresfakes

import (
	"sync"

	"github.com/smartatransit/scrapedumper/pkg/martaapi"
	"github.com/smartatransit/scrapedumper/pkg/postgres"
)

type FakeRepository struct {
	AddArrivalEstimateStub        func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) error
	addArrivalEstimateMutex       sync.RWMutex
	addArrivalEstimateArgsForCall []struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 postgres.EasternTime
		arg7 postgres.EasternTime
	}
	addArrivalEstimateReturns struct {
		result1 error
	}
	addArrivalEstimateReturnsOnCall map[int]struct {
		result1 error
	}
	CreateRunRecordStub        func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Line, martaapi.Direction, *uint, *uint) error
	createRunRecordMutex       sync.RWMutex
	createRunRecordArgsForCall []struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Line
		arg6 martaapi.Direction
		arg7 *uint
		arg8 *uint
	}
	createRunRecordReturns struct {
		result1 error
	}
	createRunRecordReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteStaleRunsStub        func(postgres.EasternTime) (int64, int64, int64, error)
	deleteStaleRunsMutex       sync.RWMutex
	deleteStaleRunsArgsForCall []struct {
		arg1 postgres.EasternTime
	}
	deleteStaleRunsReturns struct {
		result1 int64
		result2 int64
		result3 int64
		result4 error
	}
	deleteStaleRunsReturnsOnCall map[int]struct {
		result1 int64
		result2 int64
		result3 int64
		result4 error
	}
	EnsureArrivalRecordStub        func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, *uint) error
	ensureArrivalRecordMutex       sync.RWMutex
	ensureArrivalRecordArgsForCall []struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 *uint
	}
	ensureArrivalRecordReturns struct {
		result1 error
	}
	ensureArrivalRecordReturnsOnCall map[int]struct {
		result1 error
	}
	EnsureTablesStub        func(bool) error
	ensureTablesMutex       sync.RWMutex
	ensureTablesArgsForCall []struct {
		arg1 bool
	}
	ensureTablesReturns struct {
		result1 error
	}
	ensureTablesReturnsOnCall map[int]struct {
		result1 error
	}
	GetLatestEstimatesStub        func(martaapi.Station) ([]postgres.LastestEstimate, error)
	getLatestEstimatesMutex       sync.RWMutex
	getLatestEstimatesArgsForCall []struct {
		arg1 martaapi.Station
	}
	getLatestEstimatesReturns struct {
		result1 []postgres.LastestEstimate
		result2 error
	}
	getLatestEstimatesReturnsOnCall map[int]struct {
		result1 []postgres.LastestEstimate
		result2 error
	}
	GetLatestRunStartMomentForStub        func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime) (postgres.EasternTime, postgres.EasternTime, error)
	getLatestRunStartMomentForMutex       sync.RWMutex
	getLatestRunStartMomentForArgsForCall []struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
	}
	getLatestRunStartMomentForReturns struct {
		result1 postgres.EasternTime
		result2 postgres.EasternTime
		result3 error
	}
	getLatestRunStartMomentForReturnsOnCall map[int]struct {
		result1 postgres.EasternTime
		result2 postgres.EasternTime
		result3 error
	}
	GetRecentlyActiveRunsStub        func(postgres.EasternTime) (map[string]postgres.Run, error)
	getRecentlyActiveRunsMutex       sync.RWMutex
	getRecentlyActiveRunsArgsForCall []struct {
		arg1 postgres.EasternTime
	}
	getRecentlyActiveRunsReturns struct {
		result1 map[string]postgres.Run
		result2 error
	}
	getRecentlyActiveRunsReturnsOnCall map[int]struct {
		result1 map[string]postgres.Run
		result2 error
	}
	SetArrivalTimeStub        func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) error
	setArrivalTimeMutex       sync.RWMutex
	setArrivalTimeArgsForCall []struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 postgres.EasternTime
		arg7 postgres.EasternTime
	}
	setArrivalTimeReturns struct {
		result1 error
	}
	setArrivalTimeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRepository) AddArrivalEstimate(arg1 martaapi.Direction, arg2 martaapi.Line, arg3 string, arg4 postgres.EasternTime, arg5 martaapi.Station, arg6 postgres.EasternTime, arg7 postgres.EasternTime) error {
	fake.addArrivalEstimateMutex.Lock()
	ret, specificReturn := fake.addArrivalEstimateReturnsOnCall[len(fake.addArrivalEstimateArgsForCall)]
	fake.addArrivalEstimateArgsForCall = append(fake.addArrivalEstimateArgsForCall, struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 postgres.EasternTime
		arg7 postgres.EasternTime
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7})
	stub := fake.AddArrivalEstimateStub
	fakeReturns := fake.addArrivalEstimateReturns
	fake.recordInvocation("AddArrivalEstimate", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7})
	fake.addArrivalEstimateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) AddArrivalEstimateCallCount() int {
	fake.addArrivalEstimateMutex.RLock()
	defer fake.addArrivalEstimateMutex.RUnlock()
	return len(fake.addArrivalEstimateArgsForCall)
}

func (fake *FakeRepository) AddArrivalEstimateCalls(stub func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) error) {
	fake.addArrivalEstimateMutex.Lock()
	defer fake.addArrivalEstimateMutex.Unlock()
	fake.AddArrivalEstimateStub = stub
}

func (fake *FakeRepository) AddArrivalEstimateArgsForCall(i int) (martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) {
	fake.addArrivalEstimateMutex.RLock()
	defer fake.addArrivalEstimateMutex.RUnlock()
	argsForCall := fake.addArrivalEstimateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7
}

func (fake *FakeRepository) AddArrivalEstimateReturns(result1 error) {
	fake.addArrivalEstimateMutex.Lock()
	defer fake.addArrivalEstimateMutex.Unlock()
	fake.AddArrivalEstimateStub = nil
	fake.addArrivalEstimateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) AddArrivalEstimateReturnsOnCall(i int, result1 error) {
	fake.addArrivalEstimateMutex.Lock()
	defer fake.addArrivalEstimateMutex.Unlock()
	fake.AddArrivalEstimateStub = nil
	if fake.addArrivalEstimateReturnsOnCall == nil {
		fake.addArrivalEstimateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addArrivalEstimateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) CreateRunRecord(arg1 martaapi.Direction, arg2 martaapi.Line, arg3 string, arg4 postgres.EasternTime, arg5 martaapi.Line, arg6 martaapi.Direction, arg7 *uint, arg8 *uint) error {
	fake.createRunRecordMutex.Lock()
	ret, specificReturn := fake.createRunRecordReturnsOnCall[len(fake.createRunRecordArgsForCall)]
	fake.createRunRecordArgsForCall = append(fake.createRunRecordArgsForCall, struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Line
		arg6 martaapi.Direction
		arg7 *uint
		arg8 *uint
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	stub := fake.CreateRunRecordStub
	fakeReturns := fake.createRunRecordReturns
	fake.recordInvocation("CreateRunRecord", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	fake.createRunRecordMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) CreateRunRecordCallCount() int {
	fake.createRunRecordMutex.RLock()
	defer fake.createRunRecordMutex.RUnlock()
	return len(fake.createRunRecordArgsForCall)
}

func (fake *FakeRepository) CreateRunRecordCalls(stub func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Line, martaapi.Direction, *uint, *uint) error) {
	fake.createRunRecordMutex.Lock()
	defer fake.createRunRecordMutex.Unlock()
	fake.CreateRunRecordStub = stub
}

func (fake *FakeRepository) CreateRunRecordArgsForCall(i int) (martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Line, martaapi.Direction, *uint, *uint) {
	fake.createRunRecordMutex.RLock()
	defer fake.createRunRecordMutex.RUnlock()
	argsForCall := fake.createRunRecordArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7, argsForCall.arg8
}

func (fake *FakeRepository) CreateRunRecordReturns(result1 error) {
	fake.createRunRecordMutex.Lock()
	defer fake.createRunRecordMutex.Unlock()
	fake.CreateRunRecordStub = nil
	fake.createRunRecordReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) CreateRunRecordReturnsOnCall(i int, result1 error) {
	fake.createRunRecordMutex.Lock()
	defer fake.createRunRecordMutex.Unlock()
	fake.CreateRunRecordStub = nil
	if fake.createRunRecordReturnsOnCall == nil {
		fake.createRunRecordReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createRunRecordReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) DeleteStaleRuns(arg1 postgres.EasternTime) (int64, int64, int64, error) {
	fake.deleteStaleRunsMutex.Lock()
	ret, specificReturn := fake.deleteStaleRunsReturnsOnCall[len(fake.deleteStaleRunsArgsForCall)]
	fake.deleteStaleRunsArgsForCall = append(fake.deleteStaleRunsArgsForCall, struct {
		arg1 postgres.EasternTime
	}{arg1})
	stub := fake.DeleteStaleRunsStub
	fakeReturns := fake.deleteStaleRunsReturns
	fake.recordInvocation("DeleteStaleRuns", []interface{}{arg1})
	fake.deleteStaleRunsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3, ret.result4
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3, fakeReturns.result4
}

func (fake *FakeRepository) DeleteStaleRunsCallCount() int {
	fake.deleteStaleRunsMutex.RLock()
	defer fake.deleteStaleRunsMutex.RUnlock()
	return len(fake.deleteStaleRunsArgsForCall)
}

func (fake *FakeRepository) DeleteStaleRunsCalls(stub func(postgres.EasternTime) (int64, int64, int64, error)) {
	fake.deleteStaleRunsMutex.Lock()
	defer fake.deleteStaleRunsMutex.Unlock()
	fake.DeleteStaleRunsStub = stub
}

func (fake *FakeRepository) DeleteStaleRunsArgsForCall(i int) postgres.EasternTime {
	fake.deleteStaleRunsMutex.RLock()
	defer fake.deleteStaleRunsMutex.RUnlock()
	argsForCall := fake.deleteStaleRunsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) DeleteStaleRunsReturns(result1 int64, result2 int64, result3 int64, result4 error) {
	fake.deleteStaleRunsMutex.Lock()
	defer fake.deleteStaleRunsMutex.Unlock()
	fake.DeleteStaleRunsStub = nil
	fake.deleteStaleRunsReturns = struct {
		result1 int64
		result2 int64
		result3 int64
		result4 error
	}{result1, result2, result3, result4}
}

func (fake *FakeRepository) DeleteStaleRunsReturnsOnCall(i int, result1 int64, result2 int64, result3 int64, result4 error) {
	fake.deleteStaleRunsMutex.Lock()
	defer fake.deleteStaleRunsMutex.Unlock()
	fake.DeleteStaleRunsStub = nil
	if fake.deleteStaleRunsReturnsOnCall == nil {
		fake.deleteStaleRunsReturnsOnCall = make(map[int]struct {
			result1 int64
			result2 int64
			result3 int64
			result4 error
		})
	}
	fake.deleteStaleRunsReturnsOnCall[i] = struct {
		result1 int64
		result2 int64
		result3 int64
		result4 error
	}{result1, result2, result3, result4}
}

func (fake *FakeRepository) EnsureArrivalRecord(arg1 martaapi.Direction, arg2 martaapi.Line, arg3 string, arg4 postgres.EasternTime, arg5 martaapi.Station, arg6 *uint) error {
	fake.ensureArrivalRecordMutex.Lock()
	ret, specificReturn := fake.ensureArrivalRecordReturnsOnCall[len(fake.ensureArrivalRecordArgsForCall)]
	fake.ensureArrivalRecordArgsForCall = append(fake.ensureArrivalRecordArgsForCall, struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 *uint
	}{arg1, arg2, arg3, arg4, arg5, arg6})
	stub := fake.EnsureArrivalRecordStub
	fakeReturns := fake.ensureArrivalRecordReturns
	fake.recordInvocation("EnsureArrivalRecord", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6})
	fake.ensureArrivalRecordMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) EnsureArrivalRecordCallCount() int {
	fake.ensureArrivalRecordMutex.RLock()
	defer fake.ensureArrivalRecordMutex.RUnlock()
	return len(fake.ensureArrivalRecordArgsForCall)
}

func (fake *FakeRepository) EnsureArrivalRecordCalls(stub func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, *uint) error) {
	fake.ensureArrivalRecordMutex.Lock()
	defer fake.ensureArrivalRecordMutex.Unlock()
	fake.EnsureArrivalRecordStub = stub
}

func (fake *FakeRepository) EnsureArrivalRecordArgsForCall(i int) (martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, *uint) {
	fake.ensureArrivalRecordMutex.RLock()
	defer fake.ensureArrivalRecordMutex.RUnlock()
	argsForCall := fake.ensureArrivalRecordArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6
}

func (fake *FakeRepository) EnsureArrivalRecordReturns(result1 error) {
	fake.ensureArrivalRecordMutex.Lock()
	defer fake.ensureArrivalRecordMutex.Unlock()
	fake.EnsureArrivalRecordStub = nil
	fake.ensureArrivalRecordReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) EnsureArrivalRecordReturnsOnCall(i int, result1 error) {
	fake.ensureArrivalRecordMutex.Lock()
	defer fake.ensureArrivalRecordMutex.Unlock()
	fake.EnsureArrivalRecordStub = nil
	if fake.ensureArrivalRecordReturnsOnCall == nil {
		fake.ensureArrivalRecordReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.ensureArrivalRecordReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) EnsureTables(arg1 bool) error {
	fake.ensureTablesMutex.Lock()
	ret, specificReturn := fake.ensureTablesReturnsOnCall[len(fake.ensureTablesArgsForCall)]
	fake.ensureTablesArgsForCall = append(fake.ensureTablesArgsForCall, struct {
		arg1 bool
	}{arg1})
	stub := fake.EnsureTablesStub
	fakeReturns := fake.ensureTablesReturns
	fake.recordInvocation("EnsureTables", []interface{}{arg1})
	fake.ensureTablesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) EnsureTablesCallCount() int {
	fake.ensureTablesMutex.RLock()
	defer fake.ensureTablesMutex.RUnlock()
	return len(fake.ensureTablesArgsForCall)
}

func (fake *FakeRepository) EnsureTablesCalls(stub func(bool) error) {
	fake.ensureTablesMutex.Lock()
	defer fake.ensureTablesMutex.Unlock()
	fake.EnsureTablesStub = stub
}

func (fake *FakeRepository) EnsureTablesArgsForCall(i int) bool {
	fake.ensureTablesMutex.RLock()
	defer fake.ensureTablesMutex.RUnlock()
	argsForCall := fake.ensureTablesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) EnsureTablesReturns(result1 error) {
	fake.ensureTablesMutex.Lock()
	defer fake.ensureTablesMutex.Unlock()
	fake.EnsureTablesStub = nil
	fake.ensureTablesReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) EnsureTablesReturnsOnCall(i int, result1 error) {
	fake.ensureTablesMutex.Lock()
	defer fake.ensureTablesMutex.Unlock()
	fake.EnsureTablesStub = nil
	if fake.ensureTablesReturnsOnCall == nil {
		fake.ensureTablesReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.ensureTablesReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) GetLatestEstimates(arg1 martaapi.Station) ([]postgres.LastestEstimate, error) {
	fake.getLatestEstimatesMutex.Lock()
	ret, specificReturn := fake.getLatestEstimatesReturnsOnCall[len(fake.getLatestEstimatesArgsForCall)]
	fake.getLatestEstimatesArgsForCall = append(fake.getLatestEstimatesArgsForCall, struct {
		arg1 martaapi.Station
	}{arg1})
	stub := fake.GetLatestEstimatesStub
	fakeReturns := fake.getLatestEstimatesReturns
	fake.recordInvocation("GetLatestEstimates", []interface{}{arg1})
	fake.getLatestEstimatesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetLatestEstimatesCallCount() int {
	fake.getLatestEstimatesMutex.RLock()
	defer fake.getLatestEstimatesMutex.RUnlock()
	return len(fake.getLatestEstimatesArgsForCall)
}

func (fake *FakeRepository) GetLatestEstimatesCalls(stub func(martaapi.Station) ([]postgres.LastestEstimate, error)) {
	fake.getLatestEstimatesMutex.Lock()
	defer fake.getLatestEstimatesMutex.Unlock()
	fake.GetLatestEstimatesStub = stub
}

func (fake *FakeRepository) GetLatestEstimatesArgsForCall(i int) martaapi.Station {
	fake.getLatestEstimatesMutex.RLock()
	defer fake.getLatestEstimatesMutex.RUnlock()
	argsForCall := fake.getLatestEstimatesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetLatestEstimatesReturns(result1 []postgres.LastestEstimate, result2 error) {
	fake.getLatestEstimatesMutex.Lock()
	defer fake.getLatestEstimatesMutex.Unlock()
	fake.GetLatestEstimatesStub = nil
	fake.getLatestEstimatesReturns = struct {
		result1 []postgres.LastestEstimate
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetLatestEstimatesReturnsOnCall(i int, result1 []postgres.LastestEstimate, result2 error) {
	fake.getLatestEstimatesMutex.Lock()
	defer fake.getLatestEstimatesMutex.Unlock()
	fake.GetLatestEstimatesStub = nil
	if fake.getLatestEstimatesReturnsOnCall == nil {
		fake.getLatestEstimatesReturnsOnCall = make(map[int]struct {
			result1 []postgres.LastestEstimate
			result2 error
		})
	}
	fake.getLatestEstimatesReturnsOnCall[i] = struct {
		result1 []postgres.LastestEstimate
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetLatestRunStartMomentFor(arg1 martaapi.Direction, arg2 martaapi.Line, arg3 string, arg4 postgres.EasternTime) (postgres.EasternTime, postgres.EasternTime, error) {
	fake.getLatestRunStartMomentForMutex.Lock()
	ret, specificReturn := fake.getLatestRunStartMomentForReturnsOnCall[len(fake.getLatestRunStartMomentForArgsForCall)]
	fake.getLatestRunStartMomentForArgsForCall = append(fake.getLatestRunStartMomentForArgsForCall, struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
	}{arg1, arg2, arg3, arg4})
	stub := fake.GetLatestRunStartMomentForStub
	fakeReturns := fake.getLatestRunStartMomentForReturns
	fake.recordInvocation("GetLatestRunStartMomentFor", []interface{}{arg1, arg2, arg3, arg4})
	fake.getLatestRunStartMomentForMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeRepository) GetLatestRunStartMomentForCallCount() int {
	fake.getLatestRunStartMomentForMutex.RLock()
	defer fake.getLatestRunStartMomentForMutex.RUnlock()
	return len(fake.getLatestRunStartMomentForArgsForCall)
}

func (fake *FakeRepository) GetLatestRunStartMomentForCalls(stub func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime) (postgres.EasternTime, postgres.EasternTime, error)) {
	fake.getLatestRunStartMomentForMutex.Lock()
	defer fake.getLatestRunStartMomentForMutex.Unlock()
	fake.GetLatestRunStartMomentForStub = stub
}

func (fake *FakeRepository) GetLatestRunStartMomentForArgsForCall(i int) (martaapi.Direction, martaapi.Line, string, postgres.EasternTime) {
	fake.getLatestRunStartMomentForMutex.RLock()
	defer fake.getLatestRunStartMomentForMutex.RUnlock()
	argsForCall := fake.getLatestRunStartMomentForArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeRepository) GetLatestRunStartMomentForReturns(result1 postgres.EasternTime, result2 postgres.EasternTime, result3 error) {
	fake.getLatestRunStartMomentForMutex.Lock()
	defer fake.getLatestRunStartMomentForMutex.Unlock()
	fake.GetLatestRunStartMomentForStub = nil
	fake.getLatestRunStartMomentForReturns = struct {
		result1 postgres.EasternTime
		result2 postgres.EasternTime
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeRepository) GetLatestRunStartMomentForReturnsOnCall(i int, result1 postgres.EasternTime, result2 postgres.EasternTime, result3 error) {
	fake.getLatestRunStartMomentForMutex.Lock()
	defer fake.getLatestRunStartMomentForMutex.Unlock()
	fake.GetLatestRunStartMomentForStub = nil
	if fake.getLatestRunStartMomentForReturnsOnCall == nil {
		fake.getLatestRunStartMomentForReturnsOnCall = make(map[int]struct {
			result1 postgres.EasternTime
			result2 postgres.EasternTime
			result3 error
		})
	}
	fake.getLatestRunStartMomentForReturnsOnCall[i] = struct {
		result1 postgres.EasternTime
		result2 postgres.EasternTime
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeRepository) GetRecentlyActiveRuns(arg1 postgres.EasternTime) (map[string]postgres.Run, error) {
	fake.getRecentlyActiveRunsMutex.Lock()
	ret, specificReturn := fake.getRecentlyActiveRunsReturnsOnCall[len(fake.getRecentlyActiveRunsArgsForCall)]
	fake.getRecentlyActiveRunsArgsForCall = append(fake.getRecentlyActiveRunsArgsForCall, struct {
		arg1 postgres.EasternTime
	}{arg1})
	stub := fake.GetRecentlyActiveRunsStub
	fakeReturns := fake.getRecentlyActiveRunsReturns
	fake.recordInvocation("GetRecentlyActiveRuns", []interface{}{arg1})
	fake.getRecentlyActiveRunsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) GetRecentlyActiveRunsCallCount() int {
	fake.getRecentlyActiveRunsMutex.RLock()
	defer fake.getRecentlyActiveRunsMutex.RUnlock()
	return len(fake.getRecentlyActiveRunsArgsForCall)
}

func (fake *FakeRepository) GetRecentlyActiveRunsCalls(stub func(postgres.EasternTime) (map[string]postgres.Run, error)) {
	fake.getRecentlyActiveRunsMutex.Lock()
	defer fake.getRecentlyActiveRunsMutex.Unlock()
	fake.GetRecentlyActiveRunsStub = stub
}

func (fake *FakeRepository) GetRecentlyActiveRunsArgsForCall(i int) postgres.EasternTime {
	fake.getRecentlyActiveRunsMutex.RLock()
	defer fake.getRecentlyActiveRunsMutex.RUnlock()
	argsForCall := fake.getRecentlyActiveRunsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) GetRecentlyActiveRunsReturns(result1 map[string]postgres.Run, result2 error) {
	fake.getRecentlyActiveRunsMutex.Lock()
	defer fake.getRecentlyActiveRunsMutex.Unlock()
	fake.GetRecentlyActiveRunsStub = nil
	fake.getRecentlyActiveRunsReturns = struct {
		result1 map[string]postgres.Run
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) GetRecentlyActiveRunsReturnsOnCall(i int, result1 map[string]postgres.Run, result2 error) {
	fake.getRecentlyActiveRunsMutex.Lock()
	defer fake.getRecentlyActiveRunsMutex.Unlock()
	fake.GetRecentlyActiveRunsStub = nil
	if fake.getRecentlyActiveRunsReturnsOnCall == nil {
		fake.getRecentlyActiveRunsReturnsOnCall = make(map[int]struct {
			result1 map[string]postgres.Run
			result2 error
		})
	}
	fake.getRecentlyActiveRunsReturnsOnCall[i] = struct {
		result1 map[string]postgres.Run
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) SetArrivalTime(arg1 martaapi.Direction, arg2 martaapi.Line, arg3 string, arg4 postgres.EasternTime, arg5 martaapi.Station, arg6 postgres.EasternTime, arg7 postgres.EasternTime) error {
	fake.setArrivalTimeMutex.Lock()
	ret, specificReturn := fake.setArrivalTimeReturnsOnCall[len(fake.setArrivalTimeArgsForCall)]
	fake.setArrivalTimeArgsForCall = append(fake.setArrivalTimeArgsForCall, struct {
		arg1 martaapi.Direction
		arg2 martaapi.Line
		arg3 string
		arg4 postgres.EasternTime
		arg5 martaapi.Station
		arg6 postgres.EasternTime
		arg7 postgres.EasternTime
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7})
	stub := fake.SetArrivalTimeStub
	fakeReturns := fake.setArrivalTimeReturns
	fake.recordInvocation("SetArrivalTime", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7})
	fake.setArrivalTimeMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) SetArrivalTimeCallCount() int {
	fake.setArrivalTimeMutex.RLock()
	defer fake.setArrivalTimeMutex.RUnlock()
	return len(fake.setArrivalTimeArgsForCall)
}

func (fake *FakeRepository) SetArrivalTimeCalls(stub func(martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) error) {
	fake.setArrivalTimeMutex.Lock()
	defer fake.setArrivalTimeMutex.Unlock()
	fake.SetArrivalTimeStub = stub
}

func (fake *FakeRepository) SetArrivalTimeArgsForCall(i int) (martaapi.Direction, martaapi.Line, string, postgres.EasternTime, martaapi.Station, postgres.EasternTime, postgres.EasternTime) {
	fake.setArrivalTimeMutex.RLock()
	defer fake.setArrivalTimeMutex.RUnlock()
	argsForCall := fake.setArrivalTimeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7
}

func (fake *FakeRepository) SetArrivalTimeReturns(result1 error) {
	fake.setArrivalTimeMutex.Lock()
	defer fake.setArrivalTimeMutex.Unlock()
	fake.SetArrivalTimeStub = nil
	fake.setArrivalTimeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) SetArrivalTimeReturnsOnCall(i int, result1 error) {
	fake.setArrivalTimeMutex.Lock()
	defer fake.setArrivalTimeMutex.Unlock()
	fake.SetArrivalTimeStub = nil
	if fake.setArrivalTimeReturnsOnCall == nil {
		fake.setArrivalTimeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.setArrivalTimeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addArrivalEstimateMutex.RLock()
	defer fake.addArrivalEstimateMutex.RUnlock()
	fake.createRunRecordMutex.RLock()
	defer fake.createRunRecordMutex.RUnlock()
	fake.deleteStaleRunsMutex.RLock()
	defer fake.deleteStaleRunsMutex.RUnlock()
	fake.ensureArrivalRecordMutex.RLock()
	defer fake.ensureArrivalRecordMutex.RUnlock()
	fake.ensureTablesMutex.RLock()
	defer fake.ensureTablesMutex.RUnlock()
	fake.getLatestEstimatesMutex.RLock()
	defer fake.getLatestEstimatesMutex.RUnlock()
	fake.getLatestRunStartMomentForMutex.RLock()
	defer fake.getLatestRunStartMomentForMutex.RUnlock()
	fake.getRecentlyActiveRunsMutex.RLock()
	defer fake.getRecentlyActiveRunsMutex.RUnlock()
	fake.setArrivalTimeMutex.RLock()
	defer fake.setArrivalTimeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRepository) recordInvocation(key string, args []interface{}) {
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

var _ postgres.Repository = new(FakeRepository)
