package domain

import "errors"

var (
	// ErrPaymentNotFound indicates payment was not found
	ErrPaymentNotFound = errors.New("payment not found")
	
	// ErrInvalidAmount indicates invalid payment amount
	ErrInvalidAmount = errors.New("amount must be greater than 0")
	
	// ErrInvalidRequest indicates invalid request payload
	ErrInvalidRequest = errors.New("invalid request payload")
	
	// ErrPaymentAlreadyCompleted indicates payment is already finalized
	ErrPaymentAlreadyCompleted = errors.New("payment already completed")
	
	// ErrOrderIDRequired indicates order ID is missing
	ErrOrderIDRequired = errors.New("order ID is required")
	
	// ErrAuthenticationFailed indicates PayMob authentication failed
	ErrAuthenticationFailed = errors.New("PayMob authentication failed")
	
	// ErrOrderCreationFailed indicates PayMob order creation failed
	ErrOrderCreationFailed = errors.New("PayMob order creation failed")
	
	// ErrPaymentKeyFailed indicates PayMob payment key generation failed
	ErrPaymentKeyFailed = errors.New("PayMob payment key generation failed")
	
	// ErrInvalidSignature indicates webhook signature verification failed
	ErrInvalidSignature = errors.New("invalid webhook signature")
)
