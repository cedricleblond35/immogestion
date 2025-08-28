import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { catchError, timeout, retry, tap, takeUntil, Observable, throwError, timer, firstValueFrom } from 'rxjs';
import { LoginCredentials, AuthResponse } from '../models/auth.interface';
import { Subject } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly API_URL = 'http://localhost:8080/api';
  private destroy$ = new Subject<void>();

  // ===== APPEL API DE CONNEXION =====
  async callLoginAPI(credentials: LoginCredentials): Promise<AuthResponse> {
    console.log('Appel API de connexion...');
    
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      'X-Request-ID': this.generateRequestId()
    });

    const url = `${this.API_URL}/auth/login`;
    
    return firstValueFrom(
    this.http.post<AuthResponse>(url, credentials, { headers, withCredentials: true }).pipe(
      tap(response => console.log('Réponse API reçue:', response)),
      timeout(10000),
      retry({ count: 2, delay: this.retryStrategy }),
      takeUntil(this.destroy$),
      catchError(this.handleHttpError.bind(this))
    )
  );
  }

  // ===== STRATÉGIE DE RETRY =====
  private retryStrategy = (error: any, retryCount: number) => {
    console.log(`Tentative ${retryCount} après erreur:`, error.status);

    if (error.status >= 400 && error.status < 500) {
      console.log('Pas de retry pour erreur client');
      return throwError(() => error);
    }

    const delay = retryCount * 1000;
    console.log(`Retry dans ${delay}ms...`);
    return timer(delay);
  };

  private handleHttpError(error: HttpErrorResponse): Observable<never> {
    return throwError(() => error);
  }

  private generateRequestId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}