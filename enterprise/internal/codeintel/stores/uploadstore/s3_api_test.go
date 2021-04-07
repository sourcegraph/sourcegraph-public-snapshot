package uploadstore

import (
	"testing"
	"time"
)

func TestS3Lifecycle(t *testing.T) {
	if lifecycle := lifecycle(time.Hour * 24 * 3); lifecycle == nil || len(lifecycle.Rules) != 2 {
		t.Fatalf("unexpected lifecycle rules")
	} else {
		var objectExpiration *int64
		for _, rule := range lifecycle.Rules {
			if rule.Expiration != nil {
				if value := rule.Expiration.Days; value != nil {
					objectExpiration = value
				}
			}
		}
		if objectExpiration == nil {
			t.Fatalf("expected object expiration to be configured")
		} else if *objectExpiration != 3 {
			t.Errorf("unexpected ttl for object expiration. want=%d have=%d", 3, *objectExpiration)
		}

		var multipartExpiration *int64
		for _, rule := range lifecycle.Rules {
			if rule.AbortIncompleteMultipartUpload != nil {
				if value := rule.AbortIncompleteMultipartUpload.DaysAfterInitiation; value != nil {
					multipartExpiration = value
				}
			}
		}
		if multipartExpiration == nil {
			t.Fatalf("expected multipart upload expiration to be configured")
		} else if *multipartExpiration != 3 {
			t.Errorf("unexpected ttl for multipart upload expiration. want=%d have=%d", 3, *multipartExpiration)
		}
	}
}
