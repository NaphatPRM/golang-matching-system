package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NaphatPRM/golang-matching-system/internal/dispatch"
	"github.com/NaphatPRM/golang-matching-system/pkg/api"
	"github.com/NaphatPRM/golang-matching-system/pkg/services"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize services
	restaurantService := services.NewRestaurantService()
	customerService := services.NewCustomerService()
	driverService := services.NewDriverService()
	orderService := services.NewOrderService(restaurantService, customerService)

	// Initialize dispatch engine
	dispatchConfig := dispatch.DefaultDispatchConfig()
	dispatchEngine := dispatch.NewDispatchEngine(dispatchConfig, driverService, orderService)

	// Initialize API handlers
	restaurantHandler := api.NewRestaurantHandler(restaurantService)
	customerHandler := api.NewCustomerHandler(customerService)
	driverHandler := api.NewDriverHandler(driverService)
	orderHandler := api.NewOrderHandler(orderService, driverService, dispatchEngine)
	dispatchHandler := api.NewDispatchHandler(dispatchEngine)

	// Setup router
	router := mux.NewRouter()

	// Register API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	restaurantHandler.RegisterRoutes(apiRouter)
	customerHandler.RegisterRoutes(apiRouter)
	driverHandler.RegisterRoutes(apiRouter)
	orderHandler.RegisterRoutes(apiRouter)
	dispatchHandler.RegisterRoutes(apiRouter)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Add CORS middleware
	router.Use(corsMiddleware)

	// Start dispatch engine
	dispatchEngine.Start()

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("API available at http://localhost:%s/api/v1", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")

	// Stop dispatch engine
	dispatchEngine.Stop()

	log.Println("Server stopped")
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
