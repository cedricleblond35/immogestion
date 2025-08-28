import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { takeUntil} from 'rxjs/operators';
import { throwError, Subject, Observable, timer } from 'rxjs';

import { LoginCredentials, AuthResponse } from '../../models/auth.interface';
import { AuthService } from '../../services/auth';


@Component({
  selector: 'app-login',
  standalone: true,
  imports: [ReactiveFormsModule, CommonModule], // ← Reactive Forms, pas Template-driven
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login implements OnInit, OnDestroy {
  
  // ===== INJECTION DES DÉPENDANCES =====
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);
  
  // ===== SIGNALS (ANGULAR 17+) =====
  // État de l'interface utilisateur avec signals
  readonly isLoading = signal(false);
  readonly showPassword = signal(false);
  readonly errorMessage = signal('');
  readonly isBlocked = signal(false);
  readonly blockTimeRemaining = signal(0);
  readonly loginAttempts = signal(0);
  
  // Constantes
  readonly maxLoginAttempts = 5;
  
  // Computed signals (calculés automatiquement)
  readonly showError = computed(() => this.errorMessage() !== '');
  readonly formattedTimeRemaining = computed(() => {
    const remaining = this.blockTimeRemaining();
    const minutes = Math.floor(remaining / 60);
    const seconds = remaining % 60;
    return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  });
  readonly remainingAttempts = computed(() => this.maxLoginAttempts - this.loginAttempts());
  
  // ===== REACTIVE FORM (MODERNE) =====
  readonly loginForm: FormGroup = this.fb.group({
    email: [
      { value: '', disabled: false }, // ← État initial du disabled
      [
        Validators.required,
        Validators.email,
        Validators.maxLength(255)
      ]
    ],
    password: [
      { value: '', disabled: false }, // ← État initial du disabled
      [
        Validators.required,
        Validators.minLength(8),
        Validators.maxLength(128)
      ]
    ],
    rememberMe: [{ value: false, disabled: false }] // ← État initial du disabled
  });
  
  // ===== CONFIGURATION =====
  private readonly API_URL = 'http://localhost:8080/api';
  private readonly TOKEN_KEY = 'immogestion_access_token';
  private readonly REFRESH_TOKEN_KEY = 'immogestion_refresh_token';
  private readonly USER_KEY = 'immogestion_user';
  
  // ===== VARIABLES PRIVÉES =====
  private destroy$ = new Subject<void>();
  private blockTimer?: any;

  ngOnInit(): void {
    console.log('Initialisation du composant Login moderne');
    
    // Vérifier si l'utilisateur est déjà connecté
    if (this.isAuthenticated()) {
      console.log('Utilisateur déjà connecté, redirection...');
      this.router.navigate(['/dashboard']);
      return;
    }

    // Vérifier les tentatives de connexion précédentes
    this.checkLoginAttempts();
    
    // Pré-remplir l'email si "Se souvenir de moi" était coché
    const rememberedEmail = localStorage.getItem('remember_user');
    if (rememberedEmail) {
      console.log('Email mémorisé trouvé:', rememberedEmail);
      this.loginForm.patchValue({ 
        email: rememberedEmail,
        rememberMe: true 
      });
    }

    // Écouter les changements de formulaire pour validation en temps réel
    this.setupFormValidation();
  }

  ngOnDestroy(): void {
    console.log('Nettoyage du composant');
    this.destroy$.next();
    this.destroy$.complete();
    
    if (this.blockTimer) {
      clearInterval(this.blockTimer);
    }
  }

  // ===== CONFIGURATION DE LA VALIDATION EN TEMPS REEL =====
  private setupFormValidation(): void {
    // Effacer les erreurs quand l'utilisateur tape
    this.loginForm.valueChanges.pipe(
      takeUntil(this.destroy$)
    ).subscribe(() => {
      if (this.showError()) {
        this.clearError();
      }
    });

    // La gestion de l'état disabled sera faite directement dans les méthodes
    // qui changent les signals (setLoadingState, blockUser, etc.)
    // Pas besoin d'observer les signals pour cela
  }

  // ===== MÉTHODES POUR GÉRER L'ÉTAT DISABLED DU FORMULAIRE =====
  private updateFormDisabledState(disabled: boolean): void {
    const controls = ['email', 'password', 'rememberMe'];
    
    controls.forEach(controlName => {
      const control = this.loginForm.get(controlName);
      if (control) {
        if (disabled && control.enabled) {
          control.disable();
        } else if (!disabled && control.disabled) {
          control.enable();
        }
      }
    });
  }

  // Méthodes publiques pour contrôler l'état du formulaire
  private enableForm(): void {
    console.log('Activation du formulaire');
    this.updateFormDisabledState(false);
  }

  private disableForm(): void {
    console.log('Désactivation du formulaire');
    this.updateFormDisabledState(true);
  }

  // ===== SOUMISSION DU FORMULAIRE (MODERNE) =====
  async onSubmit(): Promise<void> {
    console.log('=== debut onSubmit() ===');
    
    // Marquer tous les champs comme touchés pour afficher les erreurs
    this.loginForm.markAllAsTouched();
    
    console.log('État du formulaire:', {
      valid: this.loginForm.valid,
      value: this.loginForm.value,
      errors: this.loginForm.errors
    });

    // Vérifications préalables
    if (this.isBlocked()) {
      const errorMsg = `Connexion bloquée. Réessayez dans ${Math.ceil(this.blockTimeRemaining() / 60)} minutes.`;
      console.log('Utilisateur bloqué:', errorMsg);
      this.setError(errorMsg);
      return;
    }

    if (this.loginForm.invalid) {
      console.log('Formulaire invalide');
      this.setError('Veuillez corriger les erreurs dans le formulaire');
      return;
    }

    console.log('Formulaire valide, préparation de l\'envoi...');

    // Extraction des données du formulaire (type-safe)
    const formValue = this.loginForm.getRawValue();
    const credentials: LoginCredentials = {
      email: formValue.email.toLowerCase().trim(),
      password: formValue.password
    };

    console.log('Credentials préparés:', { 
      email: credentials.email, 
      password: '[MASQUÉ]' 
    });

    // Validation supplémentaire côté client
    if (!this.validateCredentials(credentials)) {
      console.log('Validation des credentials échouée');
      return;
    }

    // Démarrage de l'appel API
    this.setLoadingState(true);

    try {
      console.log('Appel à l\'API de connexion...');
      const response = await this.authService.callLoginAPI(credentials);
      console.log('Connexion réussie:', response);
      
      await this.handleLoginSuccess(response, formValue.rememberMe);
      
    } catch (error) {
      console.log('Erreur de connexion:', error);
      this.handleLoginError(error);
      
    } finally {
      this.setLoadingState(false);
      console.log('Fin de l\'appel API');
    }
    
    console.log('=== FIN DE onSubmit() ===');
  }

  // ===== GESTION DE L'ÉTAT DE CHARGEMENT =====
  private setLoadingState(loading: boolean): void {
    this.isLoading.set(loading);
    this.clearError();
    
    // Gérer l'état disabled du formulaire
    if (loading) {
      this.disableForm();
    } else {
      // Ne réactiver que si pas bloqué
      if (!this.isBlocked()) {
        this.enableForm();
      }
    }
  }


  // ===== STRATÉGIE DE RETRY MODERNE =====
  private retryStrategy = (error: any, retryCount: number) => {
    console.log(`Tentative ${retryCount} après erreur:`, error.status);
    
    // Pas de retry pour les erreurs client (4xx)
    if (error.status >= 400 && error.status < 500) {
      console.log('Pas de retry pour erreur client');
      return throwError(() => error);
    }
    
    const delay = retryCount * 1000;
    console.log(`Retry dans ${delay}ms...`);
    return timer(delay);
  };

  // ===== GESTION DU SUCCÈS (MODERNE) =====
  private async handleLoginSuccess(response: AuthResponse, rememberMe: boolean): Promise<void> {
    console.log('Traitement du succès de connexion...');

    // Validation de la réponse
    if (!this.validateAuthResponse(response)) {
      this.setError('Réponse du serveur invalide');
      return;
    }

    // Reset des tentatives
    this.resetLoginAttempts();

    // Stockage des données d'authentification
    this.storeAuthData(response, rememberMe);

    // Actions post-connexion
    await this.performPostLoginActions(response);

    // Redirection
    const redirectUrl = this.determineRedirectUrl(response.user) || '/dashboard';
    console.log('Redirection vers:', redirectUrl);
    
    // Reset du formulaire
    this.loginForm.reset();
    
    // Navigation
    await this.router.navigate([redirectUrl]);
  }

  // ===== STOCKAGE DES DONNÉES D'AUTHENTIFICATION =====
  private storeAuthData(response: AuthResponse, rememberMe: boolean): void {
    console.log('Stockage des données d\'authentification...');
    
    // Tokens
    localStorage.setItem(this.TOKEN_KEY, response.access_token);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, response.refresh_token);
    
    // Informations utilisateur
    localStorage.setItem(this.USER_KEY, JSON.stringify(response.user));
    
    // Se souvenir de moi
    if (rememberMe) {
      localStorage.setItem('remember_user', response.user.email);
    } else {
      localStorage.removeItem('remember_user');
    }
    
    console.log('Données stockées avec succès');
  }

  // ===== ACTIONS POST-CONNEXION =====
  private async performPostLoginActions(response: AuthResponse): Promise<void> {
    console.log('Exécution des actions post-connexion...');
    
    // Métadonnées de session
    sessionStorage.setItem('login_time', new Date().toISOString());
    sessionStorage.setItem('user_role', response.user.role);
    sessionStorage.setItem('token_expires_at', 
      new Date(Date.now() + (response.expires_in * 1000)).toISOString()
    );
    
    // Initialisation basée sur le rôle
    switch (response.user.role) {
      case 'admin':
        await this.initializeAdminFeatures();
        break;
      case 'manager':
        await this.initializeManagerFeatures();
        break;
      default:
        await this.initializeUserFeatures();
    }
  }

  // ===== GESTION D'ERREUR MODERNE =====
  private handleLoginError(error: any): void {
    console.log('Gestion de l\'erreur de connexion:', error);

    // Incrémenter les tentatives avec signal
    this.loginAttempts.update(attempts => attempts + 1);
    localStorage.setItem('login_attempts', this.loginAttempts().toString());
    localStorage.setItem('last_attempt_time', Date.now().toString());

    let errorMessage = 'Erreur de connexion inconnue';

    if (error?.status === 401) {
      errorMessage = 'Email ou mot de passe incorrect';
    } else if (error?.status === 429) {
      errorMessage = 'Trop de tentatives. Veuillez patienter.';
      this.blockUser(15 * 60 * 1000);
      return;
    } else if (error?.status === 422) {
      errorMessage = 'Données invalides';
    } else if (error?.status >= 500) {
      errorMessage = 'Erreur serveur. Veuillez réessayer plus tard.';
    } else if (!navigator.onLine) {
      errorMessage = 'Pas de connexion internet';
    }

    this.setError(errorMessage);

    // Vérifier si blocage nécessaire
    if (this.loginAttempts() >= this.maxLoginAttempts) {
      this.blockUser(15 * 60 * 1000);
    } else {
      const remaining = this.maxLoginAttempts - this.loginAttempts();
      if (remaining <= 2) {
        console.warn(`${remaining} tentative(s) restante(s)`);
      }
    }

    // Effacer le mot de passe
    this.loginForm.patchValue({ password: '' });
  }

  // ===== MÉTHODES UTILITAIRES MODERNES =====

  // Gestion des erreurs avec signals
  private setError(message: string): void {
    this.errorMessage.set(message);
  }

  private clearError(): void {
    this.errorMessage.set('');
  }

  // Toggle de visibilité du mot de passe
  togglePasswordVisibility(): void {
    this.showPassword.update(show => !show);
  }

  // Validation des credentials
  private validateCredentials(credentials: LoginCredentials): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(credentials.email)) {
      this.setError('Format d\'email invalide');
      return false;
    }

    if (credentials.password.length < 8) {
      this.setError('Le mot de passe doit contenir au moins 8 caractères');
      return false;
    }

    // Protection XSS
    const dangerousChars = /<script|javascript:|on\w+\s*=/i;
    if (dangerousChars.test(credentials.email) || dangerousChars.test(credentials.password)) {
      this.setError('Caractères non autorisés détectés');
      return false;
    }

    return true;
  }

  // Validation de la réponse d'authentification
  private validateAuthResponse(response: AuthResponse): boolean {
    return !!(
      response &&
      response.access_token &&
      response.refresh_token &&
      response.expires_in &&
      response.user?.id &&
      response.user?.email
    );
  }

  // Vérification d'authentification
  private isAuthenticated(): boolean {
    const token = localStorage.getItem(this.TOKEN_KEY);
    return token !== null && !this.isTokenExpired(token);
  }

  // Vérification d'expiration du token
  private isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      return payload.exp < (currentTime + 300);
    } catch (error) {
      return true;
    }
  }

  // Génération d'ID de requête
  private generateRequestId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  // Gestion du blocage temporaire
  private blockUser(duration: number): void {
    this.isBlocked.set(true);
    this.blockTimeRemaining.set(Math.ceil(duration / 1000));
    
    // Désactiver le formulaire
    this.disableForm();
    
    this.blockTimer = setInterval(() => {
      const remaining = this.blockTimeRemaining() - 1;
      this.blockTimeRemaining.set(remaining);
      
      if (remaining <= 0) {
        this.unblockUser();
      }
    }, 1000);
    
    this.setError(
      `Trop de tentatives échouées. Réessayez dans ${Math.ceil(this.blockTimeRemaining() / 60)} minutes.`
    );
  }

  private unblockUser(): void {
    this.isBlocked.set(false);
    this.blockTimeRemaining.set(0);
    this.resetLoginAttempts();
    
    // Réactiver le formulaire seulement si pas en loading
    if (!this.isLoading()) {
      this.enableForm();
    }
    
    if (this.blockTimer) {
      clearInterval(this.blockTimer);
    }
    
    this.clearError();
  }

  // Autres méthodes utilitaires
  private checkLoginAttempts(): void {
    const attempts = localStorage.getItem('login_attempts');
    const lastAttempt = localStorage.getItem('last_attempt_time');
    
    if (attempts && lastAttempt) {
      this.loginAttempts.set(parseInt(attempts, 10));
      const lastAttemptTime = parseInt(lastAttempt, 10);
      const timeDiff = Date.now() - lastAttemptTime;
      
      if (timeDiff > 15 * 60 * 1000) {
        this.resetLoginAttempts();
      } else if (this.loginAttempts() >= this.maxLoginAttempts) {
        this.blockUser(15 * 60 * 1000 - timeDiff);
      }
    }
  }

  private resetLoginAttempts(): void {
    this.loginAttempts.set(0);
    localStorage.removeItem('login_attempts');
    localStorage.removeItem('last_attempt_time');
  }

  private determineRedirectUrl(user: any): string {
    const storedUrl = sessionStorage.getItem('redirect_url');
    if (storedUrl) return storedUrl;
    
    const roleUrls: Record<string, string> = {
      admin: '/admin/dashboard',
      manager: '/manager/dashboard',
      user: '/dashboard'
    };
    
    return roleUrls[user.role] || '/dashboard';
  }

  private handleHttpError(error: HttpErrorResponse): Observable<never> {
    return throwError(() => error);
  }

  // Méthodes d'initialisation par rôle (à implémenter)
  private async initializeAdminFeatures(): Promise<void> {
    console.log('Initialisation fonctionnalités admin');
  }

  private async initializeManagerFeatures(): Promise<void> {
    console.log('Initialisation fonctionnalités manager');
  }

  private async initializeUserFeatures(): Promise<void> {
    console.log('Initialisation fonctionnalités utilisateur');
  }

  // ===== GETTERS POUR LE TEMPLATE =====
  
  // Accès facile aux contrôles du formulaire
  get email() { return this.loginForm.get('email'); }
  get password() { return this.loginForm.get('password'); }
  get rememberMe() { return this.loginForm.get('rememberMe'); }
  
  // Vérifications de validation pour le template
  hasError(controlName: string, errorType?: string): boolean {
    const control = this.loginForm.get(controlName);
    if (!control || !control.touched) return false;
    
    return errorType ? control.hasError(errorType) : control.invalid;
  }
}