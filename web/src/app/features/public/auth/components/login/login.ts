import { Component, viewChild, OnInit, OnDestroy, inject } from '@angular/core';
import { NgForm, FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { catchError, finalize, takeUntil, timeout, retry } from 'rxjs/operators';
import { throwError, Subject, Observable, timer } from 'rxjs';

// Interfaces
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    role: string;
    firstName?: string;
    lastName?: string;
  };
}

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login implements OnInit, OnDestroy {
  
  // Injection des dépendances
  private http = inject(HttpClient);  // Utilisation de HttpClient pour les requêtes HTTP
  private router = inject(Router);  // Utilisation de Router pour la navigation
  
  // ViewChild pour le formulaire
  loginForm = viewChild<NgForm>('loginForm');
  
  // Modèle du formulaire
  loginModel = {
    email: '',
    password: '',
    rememberMe: false
  };

  // État du composant
  isLoading = false;
  showPassword = false;
  loginAttempts = 0;
  maxLoginAttempts = 5;
  isBlocked = false;
  blockTimeRemaining = 0;
  
  // Variables pour la gestion des erreurs
  errorMessage = '';
  showError = false;
  
  // Configuration
  private readonly API_URL = 'http://localhost:8080/api';
  private readonly TOKEN_KEY = 'immogestion_access_token';
  private readonly REFRESH_TOKEN_KEY = 'immogestion_refresh_token';
  private readonly USER_KEY = 'immogestion_user';
  
  // Observables
  private destroy$ = new Subject<void>();
  private blockTimer?: any;

  ngOnInit(): void {
    // Vérifier si l'utilisateur est déjà connecté
    if (this.isAuthenticated()) {
      this.router.navigate(['/dashboard']);
      return;
    }

    // Récupérer les tentatives de connexion stockées
    this.checkLoginAttempts();
    
    // Pré-remplir l'email si "Se souvenir de moi" était coché
    const rememberedEmail = localStorage.getItem('remember_user');
    if (rememberedEmail) {
      this.loginModel.email = rememberedEmail;
      this.loginModel.rememberMe = true;
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    
    if (this.blockTimer) {
      clearInterval(this.blockTimer);
    }
    
    // Nettoyer les données sensibles
    this.loginModel.password = '';
  }

  /**
   * Auto-sauvegarde
   */
  autoSave() {
    const form = this.loginForm();
    
    if (form && form.dirty && !this.isBlocked) {
      // Sauvegarder uniquement l'email (pas le mot de passe pour la sécurité)
      if (this.loginModel.email) {
        sessionStorage.setItem('draft_email', this.loginModel.email);
      }
      console.log('Auto-saving email:', this.loginModel.email);
    }
  }

  /**
   * Soumission du formulaire
   */
  async onSubmit(): Promise<void> {
    const form = this.loginForm();
    
    // Vérifier si l'utilisateur est bloqué
    if (this.isBlocked) {
      this.showErrorMessage(
        `Connexion bloquée. Réessayez dans ${Math.ceil(this.blockTimeRemaining / 60)} minutes.`
      );
      return;
    }

    // Vérifier la validité du formulaire
    if (!form?.valid) {
      this.showErrorMessage('Veuillez corriger les erreurs dans le formulaire');
      return;
    }

    // Validation des données
    if (!this.validateFormData()) {
      return;
    }

    // Préparer les credentials
    const credentials: LoginCredentials = {
      email: this.sanitizeInput(this.loginModel.email.toLowerCase().trim()),
      password: this.loginModel.password
    };

    console.log('Tentative de connexion pour:', credentials.email);

    // Démarrer le chargement
    this.isLoading = true;
    this.hideError();

    try {
      // Appel à l'API
      const response = await this.callLoginAPI(credentials);
      
      // Traiter le succès
      this.handleLoginSuccess(response);
      
    } catch (error) {
      // Traiter l'erreur
      this.handleLoginError(error);
      
    } finally {
      this.isLoading = false;
      // Nettoyer le mot de passe
      this.loginModel.password = '';
    }
  }

  /**
   * Validation des données du formulaire
   */
  private validateFormData(): boolean {
    // Validation de l'email
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(this.loginModel.email)) {
      this.showErrorMessage('Format d\'email invalide');
      return false;
    }

    // Validation du mot de passe
    if (this.loginModel.password.length < 8) {
      this.showErrorMessage('Le mot de passe doit contenir au moins 8 caractères');
      return false;
    }

    // Vérification de caractères dangereux (XSS prevention)
    const dangerousChars = /<script|javascript:|on\w+\s*=/i;
    if (dangerousChars.test(this.loginModel.email) || dangerousChars.test(this.loginModel.password)) {
      this.showErrorMessage('Caractères non autorisés détectés');
      return false;
    }

    return true;
  }

  /**
   * Appel sécurisé à l'API de connexion
   */
  private callLoginAPI(credentials: LoginCredentials): Promise<AuthResponse> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      'X-Request-ID': this.generateRequestId()  // unique ID
    });

    return this.http.post<AuthResponse>(`${this.API_URL}/auth/login`, credentials, { 
      headers,
      withCredentials: true // Pour les cookies CSRF si nécessaire
    }).pipe(
      timeout(10000), // Timeout de 10 secondes
      retry({
        count: 2,
        delay: (error, retryCount) => {
          // Retry seulement sur les erreurs réseau (pas les erreurs 4xx)
          if (error.status >= 400 && error.status < 500) {
            return throwError(() => error);
          }
          return timer(retryCount * 1000);
        }
      }),
      takeUntil(this.destroy$), //  Annuler si le composant est détruit
      catchError(this.handleHttpError.bind(this))
    ).toPromise() as Promise<AuthResponse>;
  }

  /**
   * Gestion du succès de connexion
   */
  private handleLoginSuccess(response: AuthResponse): void {
    console.log('Connexion réussie');

    // Validation de la réponse
    if (!this.validateAuthResponse(response)) {
      this.showErrorMessage('Réponse du serveur invalide');
      return;
    }

    // Reset des tentatives
    this.resetLoginAttempts();

    // Stockage sécurisé des tokens
    this.setTokens(response.access_token, response.refresh_token);
    
    // Stockage des informations utilisateur
    this.setCurrentUser(response.user);

    // Gestion du "Se souvenir de moi"
    if (this.loginModel.rememberMe) {
      localStorage.setItem('remember_user', response.user.email);
    } else {
      localStorage.removeItem('remember_user');
    }

    // Nettoyer les données de session
    sessionStorage.removeItem('draft_email');

    console.log(`Bienvenue, ${response.user.email}!`);

    // Redirection sécurisée
    const redirectUrl = this.getRedirectUrl() || '/dashboard';
    this.router.navigate([redirectUrl]);

    // Reset du formulaire
    this.resetForm();
  }

  /**
   * Gestion des erreurs de connexion
   */
  private handleLoginError(error: any): void {
    console.error('Erreur de connexion:', error);

    // Incrémenter les tentatives
    this.loginAttempts++;
    localStorage.setItem('login_attempts', this.loginAttempts.toString());
    localStorage.setItem('last_attempt_time', Date.now().toString());

    let errorMessage = 'Erreur de connexion';

    // Gestion spécifique des erreurs
    if (error?.status === 401) {
      errorMessage = 'Email ou mot de passe incorrect';
    } else if (error?.status === 429) {
      errorMessage = 'Trop de tentatives. Veuillez patienter.';
      this.blockUser(15 * 60 * 1000); // 15 minutes
      return;
    } else if (error?.status === 422) {
      errorMessage = 'Données invalides';
    } else if (error?.status >= 500) {
      errorMessage = 'Erreur serveur. Veuillez réessayer plus tard.';
    } else if (!navigator.onLine) {
      errorMessage = 'Pas de connexion internet';
    }

    this.showErrorMessage(errorMessage);

    // Vérifier si on doit bloquer l'utilisateur
    if (this.loginAttempts >= this.maxLoginAttempts) {
      this.blockUser(15 * 60 * 1000); // 15 minutes
    } else {
      const remainingAttempts = this.maxLoginAttempts - this.loginAttempts;
      if (remainingAttempts <= 2) {
        console.warn(
          `${remainingAttempts} tentative(s) restante(s) avant blocage temporaire`
        );
      }
    }
  }

  /**
   * Vérification des tentatives de connexion précédentes
   */
  private checkLoginAttempts(): void {
    const attempts = localStorage.getItem('login_attempts');
    const lastAttempt = localStorage.getItem('last_attempt_time');
    
    if (attempts && lastAttempt) {
      this.loginAttempts = parseInt(attempts, 10);
      const lastAttemptTime = parseInt(lastAttempt, 10);
      const now = Date.now();
      const timeDiff = now - lastAttemptTime;
      
      // Si plus de 15 minutes se sont écoulées, reset les tentatives
      if (timeDiff > 15 * 60 * 1000) {
        this.resetLoginAttempts();
      } else if (this.loginAttempts >= this.maxLoginAttempts) {
        this.blockUser(15 * 60 * 1000 - timeDiff);
      }
    }
  }

  /**
   * Blocage temporaire de l'utilisateur
   */
  private blockUser(remainingTime: number): void {
    this.isBlocked = true;
    this.blockTimeRemaining = Math.ceil(remainingTime / 1000);
    
    this.blockTimer = setInterval(() => {
      this.blockTimeRemaining--;
      if (this.blockTimeRemaining <= 0) {
        this.unblockUser();
      }
    }, 1000);
    
    this.showErrorMessage(
      `Trop de tentatives échouées. Réessayez dans ${Math.ceil(this.blockTimeRemaining / 60)} minutes.`
    );
  }

  /**
   * Déblocage de l'utilisateur
   */
  private unblockUser(): void {
    this.isBlocked = false;
    this.blockTimeRemaining = 0;
    this.resetLoginAttempts();
    
    if (this.blockTimer) {
      clearInterval(this.blockTimer);
    }
    
    console.log('Vous pouvez maintenant réessayer de vous connecter.');
    this.hideError();
  }

  /**
   * Reset des tentatives de connexion
   */
  private resetLoginAttempts(): void {
    this.loginAttempts = 0;
    localStorage.removeItem('login_attempts');
    localStorage.removeItem('last_attempt_time');
  }

  /**
   * Utilities pour l'authentification
   */
  private isAuthenticated(): boolean {
    const token = localStorage.getItem(this.TOKEN_KEY);
    return token !== null && !this.isTokenExpired(token);
  }

  private isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      return payload.exp < (currentTime + 300); // 5 min de marge
    } catch (error) {
      return true;
    }
  }

  // Stockage sécurisé des tokens et utilisateur
  private setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem(this.TOKEN_KEY, accessToken);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
  }

  // Stockage des informations utilisateur
  private setCurrentUser(user: any): void {
    localStorage.setItem(this.USER_KEY, JSON.stringify(user));
  }

  // Récupération de l'URL de redirection
  private getRedirectUrl(): string | null {
    return sessionStorage.getItem('redirect_url');
  }

  /**
   * Validation et sanitisation
   */
  private validateAuthResponse(response: AuthResponse): boolean {
    return !!(
      response &&
      response.access_token &&
      response.refresh_token &&
      response.expires_in &&
      response.user &&
      response.user.id &&
      response.user.email
    );
  }

  private sanitizeInput(input: string): string {
    return input
      .replace(/[<>]/g, '') // Suppression des balises HTML de base
      .substring(0, 255); // Limitation de la longueur
  }

  private generateRequestId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  private handleHttpError(error: HttpErrorResponse): Observable<never> {
    return throwError(() => error);
  }

  /**
   * Gestion de l'affichage des erreurs
   */
  private showErrorMessage(message: string): void {
    this.errorMessage = message;
    this.showError = true;
  }

  private hideError(): void {
    this.showError = false;
    this.errorMessage = '';
  }

  /**
   * Reset du formulaire
   */
  private resetForm(): void {
    this.loginModel = {
      email: '',
      password: '',
      rememberMe: false
    };
    
    const form = this.loginForm();
    if (form) {
      form.resetForm();
    }
  }

  /**
   * Méthodes utilitaires pour le template
   */
  togglePasswordVisibility(): void {
    this.showPassword = !this.showPassword;
  }

  getFormattedTimeRemaining(): string {
    const minutes = Math.floor(this.blockTimeRemaining / 60);
    const seconds = this.blockTimeRemaining % 60;
    return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }

  /**
   * Validation en temps réel pour le template
   */
  isEmailValid(): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return this.loginModel.email === '' || emailRegex.test(this.loginModel.email);
  }

  isPasswordValid(): boolean {
    return this.loginModel.password === '' || this.loginModel.password.length >= 8;
  }
}