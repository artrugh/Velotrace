package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBikeImage_PopulateURL(t *testing.T) {
	bikeID := uuid.New()

	tests := []struct {
		name        string
		objectKey   string
		baseURL     string
		bucket      string
		expectedURL string
	}{
		{
			name:        "Populates URL with all env vars set",
			objectKey:   "bikes/abc/123.jpg",
			baseURL:     "https://cdn.example.com",
			bucket:      "velotrace",
			expectedURL: "https://cdn.example.com/velotrace/bikes/abc/123.jpg",
		},
		{
			name:        "Empty ObjectKey results in no URL change",
			objectKey:   "",
			baseURL:     "https://cdn.example.com",
			bucket:      "velotrace",
			expectedURL: "",
		},
		{
			name:        "Empty base URL produces partial URL",
			objectKey:   "bikes/abc/123.jpg",
			baseURL:     "",
			bucket:      "velotrace",
			expectedURL: "/velotrace/bikes/abc/123.jpg",
		},
		{
			name:        "Empty bucket produces partial URL",
			objectKey:   "bikes/abc/123.jpg",
			baseURL:     "https://cdn.example.com",
			bucket:      "",
			expectedURL: "https://cdn.example.com//bikes/abc/123.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("STORAGE_PUBLIC_BASE_URL", tt.baseURL)
			t.Setenv("STORAGE_BUCKET", tt.bucket)

			img := &BikeImage{
				ID:        uuid.New(),
				BikeID:    bikeID,
				ObjectKey: tt.objectKey,
			}

			img.PopulateURL()

			assert.Equal(t, tt.expectedURL, img.URL)
		})
	}
}

func TestBikeImage_PopulateURL_DoesNotOverwriteObjectKey(t *testing.T) {
	t.Setenv("STORAGE_PUBLIC_BASE_URL", "https://cdn.example.com")
	t.Setenv("STORAGE_BUCKET", "velotrace")

	img := &BikeImage{
		ObjectKey: "bikes/test/photo.jpg",
	}

	img.PopulateURL()

	// ObjectKey should be unchanged after PopulateURL
	assert.Equal(t, "bikes/test/photo.jpg", img.ObjectKey)
	assert.Equal(t, "https://cdn.example.com/velotrace/bikes/test/photo.jpg", img.URL)
}

func TestBikeStatusConstants(t *testing.T) {
	assert.Equal(t, BikeStatus("registered"), StatusRegistered)
	assert.Equal(t, BikeStatus("for_sale"), StatusForSale)
	assert.Equal(t, BikeStatus("stolen"), StatusStolen)
	assert.Equal(t, BikeStatus("transferred"), StatusTransferred)
}