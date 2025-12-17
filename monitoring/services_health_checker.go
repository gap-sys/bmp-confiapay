package monitoring

// Representa um objeto que checka a saúde de um serviço.
type Checker interface {
	Check() (map[string]any, error)
	GetServiceName() string
}

// Representa um objeto que checka a saúde de vários serviços.
type ServicesHealthChecker struct {
	checkers []Checker
}

func NewServicesHealthChecker(checkers ...Checker) *ServicesHealthChecker {
	return &ServicesHealthChecker{
		checkers: checkers,
	}
}

func (h *ServicesHealthChecker) Check() (map[string]any, error) {
	var info = make(map[string]any)
	for _, service := range h.checkers {
		data, _ := service.Check()
		info[service.GetServiceName()] = data
	}
	return info, nil
}
