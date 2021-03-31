package rabbitmq

import (
	"context"
	"testing"
)

type ContentTypeTestCases struct {
	URl         string
	ContentType ContentType
}

func TestGetContentType(t *testing.T) {
	testCases := []ContentTypeTestCases{
		{
			URl:         "http://www.sample.come/file.pdf",
			ContentType: MISC_CONTENT,
		},
		{
			URl:         "http://www.sample.com/file.mp4",
			ContentType: VIDEO_CONTENT,
		},
		{
			URl:         "http://www.sample.com/file.jpeg",
			ContentType: IMAGE_CONTENT,
		},
	}
	for _, tc := range testCases {
		ctx := context.Background()
		contentType := getContentType(ctx, tc.URl)
		if contentType != tc.ContentType {
			t.Errorf("Expected %s content type. Obtained %s ", tc.ContentType, contentType)
		}
	}
}
