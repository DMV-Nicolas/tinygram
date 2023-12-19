package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/DMV-Nicolas/tinygram/db/mock"
	db "github.com/DMV-Nicolas/tinygram/db/mongo"
	"github.com/DMV-Nicolas/tinygram/token"
	"github.com/DMV-Nicolas/tinygram/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCreateLikeAPI(t *testing.T) {
	user, _ := randomUser(t)
	post := randomPost(t, primitive.NewObjectID())
	like := randomLike(t, user.ID, post.ID)
	result := &mongo.InsertOneResult{InsertedID: like.ID}

	testCases := []struct {
		name          string
		body          map[string]any
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{
				"post_id": like.PostID.Hex(),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), "_id", gomock.Eq(like.PostID)).
					Times(1).
					Return(post, nil)

				arg := db.CreateLikeParams{
					UserID: user.ID,
					PostID: post.ID,
				}

				querier.EXPECT().
					CreateLike(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(result, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchInsertOneResult(t, recorder.Body, result)
			},
		},
		{
			name: "DuplicatedLike",
			body: map[string]any{
				"post_id": like.PostID.Hex(),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), "_id", gomock.Eq(like.PostID)).
					Times(1).
					Return(post, nil)

				arg := db.CreateLikeParams{
					UserID: user.ID,
					PostID: post.ID,
				}

				querier.EXPECT().
					CreateLike(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, db.ErrDuplicatedLike)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: map[string]any{
				"post_id": like.PostID.Hex(),
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), "_id", gomock.Eq(like.PostID)).
					Times(1).
					Return(post, nil)
				querier.EXPECT().
					CreateLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, mongo.ErrClientDisconnected)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPost",
			body: map[string]any{
				"post_id": "qwertyuiopasdfghjklñzxcv",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Any, gomock.Any()).
					Times(0)
				querier.EXPECT().
					CreateLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "PostIDLenIsNot24",
			body: map[string]any{
				"post_id": ":S",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Any, gomock.Any()).
					Times(0)
				querier.EXPECT().
					CreateLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			queries := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(queries)

			// marshal data body to json
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			// start test server and send request
			server := newTestServer(t, queries)
			recorder := httptest.NewRecorder()

			url := "/likes"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			request.Header.Add("Content-Type", "application/json")
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteLikeAPI(t *testing.T) {
	user, _ := randomUser(t)
	post := randomPost(t, primitive.NewObjectID())
	like := randomLike(t, user.ID, post.ID)
	result := &mongo.DeleteResult{
		DeletedCount: 1,
	}

	testCases := []struct {
		name          string
		id            any
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   like.ID.Hex(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetLike(gomock.Any(), gomock.Eq(like.ID)).
					Times(1).
					Return(like, nil)
				querier.EXPECT().
					DeleteLike(gomock.Any(), gomock.Eq(like.ID)).
					Times(1).
					Return(result, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchDeleteResult(t, recorder.Body, result)
			},
		},
		{
			name: "InternalError",
			id:   like.ID.Hex(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetLike(gomock.Any(), gomock.Eq(like.ID)).
					Times(1).
					Return(like, nil)
				querier.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, mongo.ErrClientDisconnected)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "NonLikeOwner",
			id:   like.ID.Hex(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, primitive.NewObjectID(), time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetLike(gomock.Any(), gomock.Eq(like.ID)).
					Times(1).
					Return(like, nil)
				querier.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidLike",
			id:   like.ID.Hex(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, primitive.NewObjectID(), time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Like{}, mongo.ErrNoDocuments)
				querier.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IDLenIsNot24",
			id:   ":c",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, primitive.NewObjectID(), time.Minute)
			}, buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetLike(gomock.Any(), gomock.Any()).
					Times(0)
				querier.EXPECT().
					DeleteLike(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			queries := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(queries)

			// marshal data body to json
			data, err := json.Marshal(map[string]any{"id": tc.id})
			require.NoError(t, err)

			// start test server and send request
			server := newTestServer(t, queries)
			recorder := httptest.NewRecorder()

			url := "/likes"
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/json")

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListLikesAPI(t *testing.T) {
	offset, limit := 5, 10
	post := randomPost(t, primitive.NewObjectID())
	likes := make([]db.Like, limit-offset)
	for i := 0; i < limit-offset; i++ {
		likes[i] = randomLike(t, primitive.NewObjectID(), post.ID)
	}

	testCases := []struct {
		name          string
		query         map[string]any
		buildStubs    func(store *mockdb.MockQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: map[string]any{
				"post_id": post.ID.Hex(),
				"offset":  offset,
				"limit":   limit,
			},
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Eq("_id"), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)

				arg := db.ListLikesParams{
					PostID: post.ID,
					Offset: int64(offset),
					Limit:  int64(limit),
				}

				querier.EXPECT().
					ListLikes(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(likes, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: map[string]any{
				"post_id": post.ID.Hex(),
				"offset":  offset,
				"limit":   limit,
			},
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Eq("_id"), gomock.Eq(post.ID)).
					Times(1).
					Return(post, nil)
				querier.EXPECT().
					ListLikes(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, mongo.ErrClientDisconnected)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPost",
			query: map[string]any{
				"post_id": "qwertyuiopasdfghjklñzxcv",
				"offset":  offset,
				"limit":   limit,
			},
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				querier.EXPECT().
					ListLikes(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "PostIDLenIsNot24",
			query: map[string]any{
				"post_id": "0-0",
				"offset":  offset,
				"limit":   limit,
			},
			buildStubs: func(querier *mockdb.MockQuerier) {
				querier.EXPECT().
					GetPost(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				querier.EXPECT().
					ListLikes(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			queries := mockdb.NewMockQuerier(ctrl)
			tc.buildStubs(queries)

			// start test server and send request
			server := newTestServer(t, queries)
			recorder := httptest.NewRecorder()

			url := "/likes"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Add("Content-Type", "application/json")

			q := request.URL.Query()
			q.Add("post_id", fmt.Sprint(tc.query["post_id"]))
			q.Add("offset", fmt.Sprint(tc.query["offset"]))
			q.Add("limit", fmt.Sprint(tc.query["limit"]))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomLike(t *testing.T, userID, postID primitive.ObjectID) db.Like {
	return db.Like{
		ID:     util.RandomID(),
		UserID: userID,
		PostID: postID,
	}
}
