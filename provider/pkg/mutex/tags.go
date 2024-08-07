package mutex

import (
	"fmt"
	"sync"
)

var tagRegistry = struct {
	sync.Mutex
	arnTagLocks map[string]map[string]*sync.Mutex
	arnTags     map[string]map[string]bool
}{
	arnTagLocks: make(map[string]map[string]*sync.Mutex),
	arnTags:     make(map[string]map[string]bool),
}

// BorrowTag provides a lease on a tag for a given ARN. The returned function must be called to conclude the lease.
// If another write operation has already been registered for the tag on the ARN, BorrowTag will return an error.
func BorrowTag(arn, tag string) (func(isWriteOp bool), error) {
	tagRegistry.Lock()
	if _, ok := tagRegistry.arnTagLocks[arn]; !ok {
		tagRegistry.arnTagLocks[arn] = make(map[string]*sync.Mutex)
	}
	if _, ok := tagRegistry.arnTagLocks[arn][tag]; !ok {
		tagRegistry.arnTagLocks[arn][tag] = &sync.Mutex{}
	}
	tagRegistry.Unlock()
	tagRegistry.arnTagLocks[arn][tag].Lock()

	if _, ok := tagRegistry.arnTags[arn]; !ok {
		tagRegistry.arnTags[arn] = make(map[string]bool)
	}

	if isWriteOp, ok := tagRegistry.arnTags[arn][tag]; ok && isWriteOp {
		return nil, fmt.Errorf("another write operation has already been registered for tag %q on ARN %q", tag, arn)
	}

	return func(isWriteOp bool) {
		tagRegistry.arnTags[arn][tag] = isWriteOp
		tagRegistry.arnTagLocks[arn][tag].Lock()
	}, nil
}

func BorrowTags(arns []string, tags []string) (func(isWriteOp bool), error) {
	for _, arn := range arns {
		for _, tag := range tags {
			if _, err := BorrowTag(arn, tag); err != nil {
				return nil, err
			}
		}
	}
	return func(isWriteOp bool) {
		for _, arn := range arns {
			for _, tag := range tags {
				if _, err := BorrowTag(arn, tag); err != nil {
					return
				}
			}
		}
	}, nil
}
