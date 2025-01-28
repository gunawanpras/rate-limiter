package ratelimiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gunawanpras/rate-limiter/cache"
	cacheMock "github.com/gunawanpras/rate-limiter/cache/mock"
	"github.com/gunawanpras/rate-limiter/config"
	"github.com/gunawanpras/rate-limiter/helper"
	ratelimiter "github.com/gunawanpras/rate-limiter/rate-limiter"
)

type mockTimeNowHelper struct {
	now time.Time
}

func (m *mockTimeNowHelper) Now() time.Time {
	return m.now
}

func (m *mockTimeNowHelper) GetElapsedTime(finishTime time.Time) time.Duration {
	return m.now.Sub(finishTime)
}

type mockReqHelper struct {
	ip string
}

func (m *mockReqHelper) GetIp(r *http.Request) string {
	return m.ip
}

var (
	ctx  = context.Background()
	conf = config.Config{
		RateLimiter: config.RateLimiter{
			Limit:    3,
			Interval: 30,
		},
		Cache: cache.CacheConfig{
			Ttl: 5,
		},
	}
	ip             = "192.168.1.1"
	errHappen      = errors.New("error happened")
	errKeyNotFound = errors.New("cache: key is missing")
)

func TestRateLimiter_Allow(t *testing.T) {
	type fields struct {
		CacheMock *cacheMock.MockICache
	}

	type args struct {
		ctx context.Context
		ip  string
	}

	helper.TimeHelper = &mockTimeNowHelper{
		now: time.Date(2025, time.January, 27, 18, 0, 0, 0, time.UTC),
	}

	visitor := ratelimiter.Visitor{
		LastSeen: helper.TimeHelper.Now(),
		Count:    1,
	}

	visitorBytes, _ := json.Marshal(visitor)

	tests := []struct {
		name      string
		args      args
		mockFn    func(f *fields)
		want      bool
		wantError bool
	}{
		{
			name: "Error while getting value from cache",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				f.CacheMock.EXPECT().GetValue(ctx, ip).Return("", errHappen).Times(1)
			},
			want:      false,
			wantError: true,
		},
		{
			name: "Error while setting value to cache",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				f.CacheMock.EXPECT().GetValue(ctx, ip).Return("", errKeyNotFound)
				f.CacheMock.EXPECT().SetValue(ctx, ip, visitorBytes, time.Minute*5).Return(errHappen).Times(1)
			},
			want:      false,
			wantError: true,
		},
		{
			name: "First request should be allowed",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				f.CacheMock.EXPECT().GetValue(ctx, ip).Return("", errKeyNotFound).Times(1)
				f.CacheMock.EXPECT().SetValue(ctx, ip, visitorBytes, time.Minute*5).Return(nil).Times(1)
			},
			want:      true,
			wantError: false,
		},
		{
			name: "Rate limit not exceeded, but error happen while setting value to cache",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:02Z","Count":1}`

				f.CacheMock.EXPECT().GetValue(ctx, ip).Return(getValueResponse, nil).Times(1)

				var storedVisitor ratelimiter.Visitor
				_ = json.Unmarshal([]byte(getValueResponse), &storedVisitor)

				storedVisitor.LastSeen = helper.TimeHelper.Now()
				storedVisitor.Count += 1

				visitorBytes, _ := json.Marshal(storedVisitor)

				f.CacheMock.EXPECT().SetValue(ctx, ip, visitorBytes, time.Minute*5).Return(errHappen).Times(1)
			},
			want:      false,
			wantError: true,
		},
		{
			name: "Return true while rate limit not exceeded",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:03Z","Count":2}`

				f.CacheMock.EXPECT().GetValue(ctx, ip).Return(getValueResponse, nil).Times(1)

				var storedVisitor ratelimiter.Visitor
				_ = json.Unmarshal([]byte(getValueResponse), &storedVisitor)

				storedVisitor.LastSeen = helper.TimeHelper.Now()
				storedVisitor.Count += 1

				visitorBytes, _ := json.Marshal(storedVisitor)

				f.CacheMock.EXPECT().SetValue(ctx, ip, visitorBytes, time.Minute*5).Return(nil).Times(1)
			},
			want:      true,
			wantError: false,
		},
		{
			name: "Return false while rate limit exceeded",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:03Z","Count":3}`
				f.CacheMock.EXPECT().GetValue(ctx, ip).Return(getValueResponse, nil).Times(1)
			},
			want:      false,
			wantError: false,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := cacheMock.NewMockICache(ctrl)
	ttfields := &fields{
		CacheMock: mockCache,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockFn != nil {
				tt.mockFn(ttfields)
			}

			rateLimiter := ratelimiter.NewRateLimiter(conf, mockCache)

			got, err := rateLimiter.Allow(tt.args.ip)
			if (err != nil) != tt.wantError {
				t.Errorf("Allow() error = %v, want %v", err, tt.wantError)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Allow() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateLimiter_Handler(t *testing.T) {
	type fields struct {
		CacheMock *cacheMock.MockICache
	}

	type args struct {
		ctx context.Context
		ip  string
	}

	ctx := context.Background()
	helper.ReqHelper = &mockReqHelper{
		ip: ip,
	}

	tests := []struct {
		name      string
		args      args
		mockFn    func(f *fields)
		want      int
		wantError bool
	}{
		{
			name: "Success",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:02Z","Count":1}`
				f.CacheMock.EXPECT().GetValue(context.Background(), ip).Return(getValueResponse, nil).Times(1)
				f.CacheMock.EXPECT().SetValue(context.Background(), ip, gomock.Any(), time.Minute*5).Times(1)
			},
			want:      http.StatusOK,
			wantError: false,
		},
		{
			name: "Error while unmarshal",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:02Z","Count":}`
				f.CacheMock.EXPECT().GetValue(context.Background(), ip).Return(getValueResponse, nil).Times(1)
			},
			want:      http.StatusInternalServerError,
			wantError: false,
		},
		{
			name: "Error when rate limit exceeded",
			args: args{
				ctx: ctx,
				ip:  ip,
			},
			mockFn: func(f *fields) {
				getValueResponse := `{"LastSeen":"2025-01-27T18:00:03Z","Count":3}`
				f.CacheMock.EXPECT().GetValue(context.Background(), ip).Return(getValueResponse, nil).Times(1)
			},
			want:      http.StatusTooManyRequests,
			wantError: false,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := cacheMock.NewMockICache(ctrl)
	handler := ratelimiter.NewRateLimiter(conf, mockCache).Handler()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	ttfields := &fields{
		CacheMock: mockCache,
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockFn != nil {
				tt.mockFn(ttfields)
			}

			req.RemoteAddr = tt.args.ip
			got, err := (&http.Client{}).Do(req)
			got.Body.Close()

			if (err != nil) != tt.wantError {
				t.Errorf("Handler() error = %v, want %v", err, tt.wantError)
			}

			if !reflect.DeepEqual(got.StatusCode, tt.want) {
				t.Errorf("Handler() got = %v, want %v", got.StatusCode, tt.want)
			}
		})
	}
}
