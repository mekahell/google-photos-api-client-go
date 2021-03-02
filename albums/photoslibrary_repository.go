package albums

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gphotosuploader/googlemirror/api/photoslibrary/v1"
)

// PhotosLibraryClient represents an albums using `gphotosuploader/googlemirror/api/photoslibrary`.
type PhotosLibraryClient interface {
	BatchAddMediaItems(albumId string, albumbatchaddmediaitemsrequest *photoslibrary.AlbumBatchAddMediaItemsRequest) *photoslibrary.AlbumBatchAddMediaItemsCall
	Create(createalbumrequest *photoslibrary.CreateAlbumRequest) *photoslibrary.AlbumsCreateCall
	Get(albumId string) *photoslibrary.AlbumsGetCall
	List() *photoslibrary.AlbumsListCall
}

// PhotosLibraryAlbumsRepository represents an albums Google Photos repository.
type PhotosLibraryAlbumsRepository struct {
	service  PhotosLibraryClient
	basePath string
}

// NewPhotosLibraryClient returns a Repository using PhotosLibrary service.
func NewPhotosLibraryClient(authenticatedClient *http.Client) (*PhotosLibraryAlbumsRepository, error) {
	return NewPhotosLibraryClientWithURL(authenticatedClient, "")
}

// NewPhotosLibraryClientWithURL returns a Repository using PhotosLibrary service with a custom URL.
func NewPhotosLibraryClientWithURL(authenticatedClient *http.Client, url string) (*PhotosLibraryAlbumsRepository, error) {
	s, err := photoslibrary.New(authenticatedClient)
	if err != nil {
		return nil, err
	}
	if url != "" {
		s.BasePath = url
	}
	return &PhotosLibraryAlbumsRepository{
		service:  photoslibrary.NewAlbumsService(s),
		basePath: s.BasePath,
	}, nil
}

// URL returns the albums repository url.
func (r PhotosLibraryAlbumsRepository) URL() string {
	return r.basePath
}

// AddManyItems adds multiple media item(s) to the specified album.
func (r PhotosLibraryAlbumsRepository) AddManyItems(ctx context.Context, albumId string, mediaItemIds []string) error {
	req := &photoslibrary.AlbumBatchAddMediaItemsRequest{
		MediaItemIds: mediaItemIds,
	}
	_, err := r.service.BatchAddMediaItems(albumId, req).Context(ctx).Do()
	return err
}

// RemoveManyItems removes multiple media item(s) from the specified album.
func (r PhotosLibraryAlbumsRepository) RemoveManyItems(ctx context.Context, albumId string, mediaItemIds []string) error {
	panic("not implemented on google mirror library")
}

// Create adds and caches a new album to the repo.
func (r PhotosLibraryAlbumsRepository) Create(ctx context.Context, title string) (*Album, error) {
	req := &photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{Title: title},
	}
	res, err := r.service.Create(req).Context(ctx).Do()
	if err != nil {
		return &NullAlbum, err
	}
	album := r.convertPhotosLibraryAlbumToAlbum(res)
	return &album, nil
}

// Get fetches and caches an album from the repo by id.
func (r PhotosLibraryAlbumsRepository) Get(ctx context.Context, albumId string) (*Album, error) {
	res, err := r.service.Get(albumId).Context(ctx).Do()
	if err != nil {
		return &NullAlbum, ErrAlbumNotFound
	}
	album := r.convertPhotosLibraryAlbumToAlbum(res)
	return &album, nil
}

// ListAll fetches and caches all the albums from the repo.
func (r PhotosLibraryAlbumsRepository) ListAll(ctx context.Context) ([]Album, error) {
	return r.ListAllWithOptions(ctx, ListOptions{ExcludeNonAppCreatedData: true})
}

// ListAllWithOptions fetches and caches all the albums from the repo.
func (r PhotosLibraryAlbumsRepository) ListAllWithOptions(ctx context.Context, params ListOptions) ([]Album, error) {
	albumsResult := make([]Album, 0)
	listFn := r.service.List
	if params.ExcludeNonAppCreatedData {
		listFn = listFn().ExcludeNonAppCreatedData
	}
	err := listFn().Pages(ctx, func(response *photoslibrary.ListAlbumsResponse) error {
		for _, res := range response.Albums {
			albumsResult = append(albumsResult, r.convertPhotosLibraryAlbumToAlbum(res))
		}
		return nil
	})
	return albumsResult, err
}

// GetByTitle fetches and caches an album from the repo by title.
func (r PhotosLibraryAlbumsRepository) GetByTitle(ctx context.Context, title string) (*Album, error) {
	ErrAlbumWasFound := fmt.Errorf("album was found")
	var result *Album
	if err := r.service.List().ExcludeNonAppCreatedData().Pages(ctx, func(response *photoslibrary.ListAlbumsResponse) error {
		if album, found := r.findByTitle(title, response.Albums); found {
			result = album
			return ErrAlbumWasFound
		}
		return nil
	}); err == ErrAlbumWasFound {
		return result, nil
	}
	return &NullAlbum, ErrAlbumNotFound
}

func (r PhotosLibraryAlbumsRepository) findByTitle(title string, albums []*photoslibrary.Album) (*Album, bool) {
	for _, a := range albums {
		if a.Title == title {
			album := r.convertPhotosLibraryAlbumToAlbum(a)
			return &album, true
		}
	}
	return &NullAlbum, false
}

func (r PhotosLibraryAlbumsRepository) convertPhotosLibraryAlbumToAlbum(a *photoslibrary.Album) Album {
	return Album{
		ID:                    a.Id,
		Title:                 a.Title,
		ProductURL:            a.ProductUrl,
		IsWriteable:           a.IsWriteable,
		MediaItemsCount:       strconv.FormatInt(a.TotalMediaItems, 10),
		CoverPhotoBaseURL:     a.CoverPhotoBaseUrl,
		CoverPhotoMediaItemID: "",
	}
}
