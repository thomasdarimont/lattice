// This file was generated by counterfeiter
package fake_target_verifier

import (
	"sync"

	"github.com/cloudfoundry-incubator/lattice/ltc/config/target_verifier"
)

type FakeTargetVerifier struct {
	VerifyTargetStub        func(name string) (up bool, auth bool, err error)
	verifyTargetMutex       sync.RWMutex
	verifyTargetArgsForCall []struct {
		name string
	}
	verifyTargetReturns struct {
		result1 bool
		result2 bool
		result3 error
	}
	VerifyBlobTargetStub        func(host string, port uint16, accessKey, secretKey, bucketName string) (ok bool, err error)
	verifyBlobTargetMutex       sync.RWMutex
	verifyBlobTargetArgsForCall []struct {
		host       string
		port       uint16
		accessKey  string
		secretKey  string
		bucketName string
	}
	verifyBlobTargetReturns struct {
		result1 bool
		result2 error
	}
}

func (fake *FakeTargetVerifier) VerifyTarget(name string) (up bool, auth bool, err error) {
	fake.verifyTargetMutex.Lock()
	fake.verifyTargetArgsForCall = append(fake.verifyTargetArgsForCall, struct {
		name string
	}{name})
	fake.verifyTargetMutex.Unlock()
	if fake.VerifyTargetStub != nil {
		return fake.VerifyTargetStub(name)
	} else {
		return fake.verifyTargetReturns.result1, fake.verifyTargetReturns.result2, fake.verifyTargetReturns.result3
	}
}

func (fake *FakeTargetVerifier) VerifyTargetCallCount() int {
	fake.verifyTargetMutex.RLock()
	defer fake.verifyTargetMutex.RUnlock()
	return len(fake.verifyTargetArgsForCall)
}

func (fake *FakeTargetVerifier) VerifyTargetArgsForCall(i int) string {
	fake.verifyTargetMutex.RLock()
	defer fake.verifyTargetMutex.RUnlock()
	return fake.verifyTargetArgsForCall[i].name
}

func (fake *FakeTargetVerifier) VerifyTargetReturns(result1 bool, result2 bool, result3 error) {
	fake.VerifyTargetStub = nil
	fake.verifyTargetReturns = struct {
		result1 bool
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeTargetVerifier) VerifyBlobTarget(host string, port uint16, accessKey string, secretKey string, bucketName string) (ok bool, err error) {
	fake.verifyBlobTargetMutex.Lock()
	fake.verifyBlobTargetArgsForCall = append(fake.verifyBlobTargetArgsForCall, struct {
		host       string
		port       uint16
		accessKey  string
		secretKey  string
		bucketName string
	}{host, port, accessKey, secretKey, bucketName})
	fake.verifyBlobTargetMutex.Unlock()
	if fake.VerifyBlobTargetStub != nil {
		return fake.VerifyBlobTargetStub(host, port, accessKey, secretKey, bucketName)
	} else {
		return fake.verifyBlobTargetReturns.result1, fake.verifyBlobTargetReturns.result2
	}
}

func (fake *FakeTargetVerifier) VerifyBlobTargetCallCount() int {
	fake.verifyBlobTargetMutex.RLock()
	defer fake.verifyBlobTargetMutex.RUnlock()
	return len(fake.verifyBlobTargetArgsForCall)
}

func (fake *FakeTargetVerifier) VerifyBlobTargetArgsForCall(i int) (string, uint16, string, string, string) {
	fake.verifyBlobTargetMutex.RLock()
	defer fake.verifyBlobTargetMutex.RUnlock()
	return fake.verifyBlobTargetArgsForCall[i].host, fake.verifyBlobTargetArgsForCall[i].port, fake.verifyBlobTargetArgsForCall[i].accessKey, fake.verifyBlobTargetArgsForCall[i].secretKey, fake.verifyBlobTargetArgsForCall[i].bucketName
}

func (fake *FakeTargetVerifier) VerifyBlobTargetReturns(result1 bool, result2 error) {
	fake.VerifyBlobTargetStub = nil
	fake.verifyBlobTargetReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

var _ target_verifier.TargetVerifier = new(FakeTargetVerifier)
