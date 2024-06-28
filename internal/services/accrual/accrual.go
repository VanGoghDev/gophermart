package accrual

type AccrualFetcher struct{}

func New() *AccrualFetcher {
	return &AccrualFetcher{}
}

// Апи и нейминг пока под большим вопросом пока думаю так попробовать.

func (a AccrualFetcher) Serve() {

}

// Fetch опрашивает Систему расчета баллов лояльности.
func (a AccrualFetcher) Fetch() {

}

// Save сохраняет заказ в систему расчета баллов лояльности.
func (a AccrualFetcher) Save() {

}
