package dispatcher

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/VanGoghDev/gophermart/internal/domain/models"
	"github.com/VanGoghDev/gophermart/internal/services/accrual/client"
	"github.com/VanGoghDev/gophermart/internal/storage"
	"golang.org/x/sync/errgroup"
)

type Dispatcher struct {
	client *client.Client
	s      *storage.Storage

	log          *slog.Logger
	mu           sync.Mutex
	workersCount int32
}

func New(log *slog.Logger, strg *storage.Storage, accrlHost string, workersCount int32) *Dispatcher {
	clnt := client.New(http.Client{}, accrlHost)
	d := &Dispatcher{
		log:          log,
		s:            strg,
		client:       clnt,
		workersCount: workersCount,
	}
	return d
}

func (d *Dispatcher) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	g *errgroup.Group,
	ordersCh chan models.Order,
) error {
	defer wg.Done()
	notifyCh := make(chan time.Duration)
	waitCh := make(chan time.Time, d.workersCount)
	blackList := make(map[int32]int32, d.workersCount)

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()
		for {
			select {
			case retryDuration := <-notifyCh:
				d.mu.Lock()
				if len(waitCh) == 0 {
					for range d.workersCount {
						waitCh <- time.Now().Add(retryDuration)
					}
				}
				d.mu.Unlock()
			default:
				if len(blackList) == int(d.workersCount) {
					d.mu.Lock()
					blackList = make(map[int32]int32, d.workersCount)
					d.mu.Unlock()
				}
				continue
			}
		}
	})

	for id := range d.workersCount {
		id := id
		wg.Add(1)
		g.Go(func() error {
			defer wg.Done()
			for {
				err := d.trySendRequest(ctx, notifyCh, waitCh, ordersCh, blackList, id)
				if err != nil {
					return fmt.Errorf("%w: failed to send request", err)
				}
			}
		})
	}
	return nil
}

func (d *Dispatcher) trySendRequest(
	ctx context.Context,
	notifyCh chan time.Duration,
	waitCh chan time.Time,
	ordersCh chan models.Order,
	blackList map[int32]int32,
	workerID int32,
) error {
	// Если горутина уже запрашивала таймер, то просто скипаем.
	d.mu.Lock()
	_, ok := blackList[workerID]
	d.mu.Unlock()

	if ok {
		return nil
	}

	select {
	case sleepUntil := <-waitCh:
		d.mu.Lock()
		blackList[workerID] = workerID
		d.mu.Unlock()

		if time.Now().Before(sleepUntil) {
			sleepFor := time.Until(sleepUntil)
			d.log.DebugContext(ctx, "goroutine should sleep",
				"workerID", workerID,
				"sleepUntil", sleepUntil,
				"sleepFor", sleepFor,
			)
			time.Sleep(sleepFor)
		}
	default:

		for order := range ordersCh {
			accrl, timeout, err := d.client.GetAccrual(ctx, order.Number)
			if err != nil {
				d.log.ErrorContext(ctx, "%w: failed to get accrual", "order.Number", order.Number, "workerID", workerID)
			}
			if timeout > 0 {
				notifyCh <- timeout
			}

			err = d.s.UpdateStatusAndBalance(ctx, accrl)
			if err != nil {
				d.log.ErrorContext(ctx, "%w: failed to update order status and balance",
					"order.Number", order.Number,
					"workerID", workerID,
				)
				return fmt.Errorf("%w: failed to update order status and balance in storage", err)
			}
		}
	}
	return nil
}

func SendRequest(ctx context.Context) (timeout time.Duration) {
	return time.Second
}
