import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { catchError, timeout, retry, tap, takeUntil, Observable, throwError, timer, firstValueFrom } from 'rxjs';
import { LoginCredentials, AuthResponse } from '../models/auth.interface';
import { Subject } from 'rxjs';

import { environment } from '@environments/environment';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly API_URL = environment.apiUrl;
  private readonly API_VERSION = environment.apiVersion;
  private destroy$ = new Subject<void>();

  /**
   * Call registration API endpoint
   * @param credentials User registration credentials
   * @returns Promise with authentication response
   */
  async callRegisterAPI(credentials: LoginCredentials): Promise<AuthResponse> {
    console.log('Calling registration API...');

    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      'X-Request-ID': this.generateRequestId()
    });

    const url = `${this.API_URL}/${this.API_VERSION}/auth/register`;

    return firstValueFrom(
      this.http.post<AuthResponse>(url, credentials, { headers, withCredentials: true }).pipe(
        tap(response => console.log('API response received:', response)),
        timeout(environment.timeout),
        retry({ count: environment.retryAttempts, delay: this.retryStrategy }),
        takeUntil(this.destroy$),
        catchError(this.handleHttpError.bind(this))
      )
    );
  }

  /**
   * Call login API endpoint
   * @param credentials User login credentials
   * @returns Promise with authentication response
   */
  async callLoginAPI(credentials: LoginCredentials): Promise<AuthResponse> {
    console.log('Calling login API...');
    
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      'X-Request-ID': this.generateRequestId()
    });

    const url = `${this.API_URL}/${this.API_VERSION}/auth/login`;
    
    return firstValueFrom(
      this.http.post<AuthResponse>(url, credentials, { headers, withCredentials: true }).pipe(
        tap(response => console.log('API response received:', response)),
        timeout(environment.timeout),
        retry({ count: environment.retryAttempts, delay: this.retryStrategy }),
        takeUntil(this.destroy$),
        catchError(this.handleHttpError.bind(this))
      )
    );
  }

  /**
   * Retry strategy for failed HTTP requests
   * Implements exponential backoff for server errors
   * Client errors (4xx) are not retried
   */
  private retryStrategy = (error: any, retryCount: number) => {
    console.log(`Retry attempt ${retryCount} after error:`, error.status);

    // Don't retry client errors (4xx)
    if (error.status >= 400 && error.status < 500) {
      console.log('No retry for client error');
      return throwError(() => error);
    }

    // Exponential backoff: delay increases with each retry
    const delay = retryCount * 1000;
    console.log(`Retrying in ${delay}ms...`);
    return timer(delay);
  };

  /**
   * Handle HTTP errors and pass them through
   * @param error HTTP error response
   * @returns Observable that throws the error
   */
  private handleHttpError(error: HttpErrorResponse): Observable<never> {
    return throwError(() => error);
  }

  /**
   * Generate unique request ID for tracking
   * @returns Unique string identifier
   */
  private generateRequestId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  /**
   * Cleanup method called when service is destroyed
   * Completes all pending observables to prevent memory leaks
   */
  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}