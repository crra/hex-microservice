package memory

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"sync"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

const arbitraryString = "arbitrary"

func new() *memoryRepository {
	return &memoryRepository{
		memory: make(map[string]redirect),
		m:      sync.RWMutex{},
	}
}

func TestNew(t *testing.T) {
	_, err := New(nil, "")
	assert.NoError(t, err, "should not create any error")
}

func TestLookupFindNonExisting(t *testing.T) {
	r, err := New(nil, "")
	_, err = r.LookupFind(arbitraryString)

	assert.ErrorIs(t, err, lookup.ErrNotFound)
}

func TestLookupDeleteFindNonExisting(t *testing.T) {
	r, err := New(nil, "")
	_, err = r.DeleteFind(arbitraryString)

	assert.ErrorIs(t, err, deleter.ErrNotFound)
}

func TestDeleteNonExisting(t *testing.T) {
	r, err := New(nil, "")
	err = r.Delete(arbitraryString, arbitraryString)

	assert.ErrorIs(t, err, deleter.ErrNotFound)
}

func TestStore(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	assert.NoError(t, err)
}

func TestStoreCodeTwice(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	assert.NoError(t, err)
	rs2 := adder.RedirectStorage{}
	err = r.Store(rs2)
	assert.ErrorIs(t, err, adder.ErrDuplicate)
}

func TestStoreLookupFind(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	if assert.NoError(t, err) {
		lookupFound, lookupErr := r.LookupFind(rs.Code)
		deleterFound, deleterErr := r.DeleteFind(rs.Code)
		if assert.NoError(t, lookupErr) && assert.NoError(t, deleterErr) {

			var lookupExpected lookup.RedirectStorage
			if err := copier.Copy(&lookupExpected, &lookupFound); err != nil {
				assert.Fail(t, "error copying")
			}
			assert.Equal(t, lookupExpected, lookupFound)

			var deleterExpected deleter.RedirectStorage
			if err := copier.Copy(&deleterExpected, &deleterFound); err != nil {
				assert.Fail(t, "error copying")
			}

			assert.Equal(t, deleterExpected, deleterFound)
		}
	}
}

func TestDelete(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	if assert.NoError(t, err) {
		assert.NoError(t, r.Delete(rs.Code, rs.Token))
	}
}

func TestDeleteInvalidCode(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	if assert.NoError(t, err) {
		err = r.Delete(arbitraryString, rs.Token)
		assert.ErrorIs(t, err, deleter.ErrNotFound)
	}
}

func TestDeleteInvalidToken(t *testing.T) {
	r, err := New(nil, "")
	rs := adder.RedirectStorage{}
	err = r.Store(rs)
	if assert.NoError(t, err) {
		err = r.Delete(rs.Code, arbitraryString)
		if assert.ErrorIs(t, err, deleter.ErrNotFound) {
			assert.Contains(t, err.Error(), "invalid token")
		}
	}
}

func (r *memoryRepository) storage() map[string]redirect {
	return r.memory
}

func TestConcurrent(t *testing.T) {
	numberOfWriters := 30
	numberOfReaders := 10

	r := new()

	add := func(wg *sync.WaitGroup, i int) {
		defer wg.Done()

		rs := adder.RedirectStorage{
			Code: fmt.Sprintf("Index-%d", i),
		}
		r.Store(rs)
	}

	get := func(wg *sync.WaitGroup, i int) {
		defer wg.Done()

		// ignore non existant
		_, _ = r.findByCode(fmt.Sprintf("Index-%d", i))
	}

	var wg sync.WaitGroup

	for i := 0; i < numberOfWriters; i++ {
		wg.Add(1)
		go add(&wg, i)
	}

	for i := 0; i < numberOfReaders; i++ {
		wg.Add(1)
		go get(&wg, i)
	}

	wg.Wait()

	assert.Equal(t, numberOfWriters, len(r.memory))
}
