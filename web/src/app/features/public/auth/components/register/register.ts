import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, AbstractControl, ValidationErrors } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';

import { RegisterAPIPayload, AuthResponse } from '../../models/auth.interface';
import { AuthService } from '../../services/auth';
import { Logger } from '@core/logging/logger';
import { environment } from '@environments/environment';

/**
 * UUID v4 generator integrated in component
 * Generates a RFC4122 version 4 UUID
 * Note: For production use, consider using a well-tested library like 'uuid' for better randomness and compliance
 */
function uuidv4(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

/**
 * Custom validator for password confirmation
 * Ensures that password and confirmPassword fields match
 */
function passwordMatchValidator(control: AbstractControl): ValidationErrors | null {
  const password = control.get('password')?.value;
  const confirmPassword = control.get('confirmPassword')?.value;
  return password && confirmPassword && password === confirmPassword ? null : { passwordMismatch: true };
}

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [ReactiveFormsModule, CommonModule],
  templateUrl: './register.html',
  styleUrl: './register.scss',
})
export class Register implements OnInit, OnDestroy {
  // ===== Dependency injection =====
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);
  private readonly logger = inject(Logger);

  // ===== Lifecycle hooks =====
  ngOnInit(): void {
    this.logger.info('RegisterComponent initialized');
  }

  ngOnDestroy(): void {
    this.logger.info('RegisterComponent destroyed');
  }

  // ===== Signals for state management =====
  readonly isLoading = signal(false);
  readonly showPassword = signal(false);
  readonly errorMessage = signal('');
  readonly isBlocked = signal(false);
  readonly blockTimeRemaining = signal(0);
  readonly loginAttempts = signal(0);
  readonly submitError = signal('');

  // ===== REACTIVE FORM =====
  readonly registerForm: FormGroup = this.fb.group({
    formId: [this.generateUniqueId(), [Validators.required]],
    company: ['', [Validators.required, Validators.minLength(2)]],
    lastname: ['', [Validators.required, Validators.minLength(2)]],
    firstname: ['', [Validators.required, Validators.minLength(2)]],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
    confirmPassword: ['', [Validators.required]],
    terms: [false, [Validators.requiredTrue]],
  }, { validators: passwordMatchValidator });

  // ===== Getters for easier access to form controls in template =====
  get company() { return this.registerForm.get('company'); }
  get lastname() { return this.registerForm.get('lastname'); }
  get firstname() { return this.registerForm.get('firstname'); }
  get email() { return this.registerForm.get('email'); }
  get password() { return this.registerForm.get('password'); }
  get confirmPassword() { return this.registerForm.get('confirmPassword'); }
  get terms() { return this.registerForm.get('terms'); }

  /**
   * Generate a unique form ID
   * @returns Unique form identifier string
   */
  private generateUniqueId(): string {
    return 'form-' + uuidv4();
  }

  /**
   * Get localized error messages for form validation
   * @param controlName Name of the form control to check
   * @returns Localized error message or empty string
   */
  getErrorMessage(controlName: string): string {
    const control = this.registerForm.get(controlName);
    if (!control || !control.errors || !control.touched) {
      return '';
    }

    const errors = control.errors;
    if (errors['required']) {
      return 'This field is required';
    }
    if (errors['requiredTrue']) {
      return 'You must accept the terms';
    }
    if (errors['email']) {
      return 'Invalid email format';
    }
    if (errors['minlength']) {
      const requiredLength = errors['minlength'].requiredLength;
      return `Minimum ${requiredLength} characters required`;
    }
    if (errors['passwordMismatch']) {
      return 'Passwords do not match';
    }

    return 'Validation error';
  }

  /**
   * Handle form submission
   * Validates form, calls registration API, and handles responses/errors
   */
  async onSubmit(): Promise<void> {
    this.logger.info('=== Starting onSubmit() ===');

    // Mark all fields as touched to display validation errors
    this.registerForm.markAllAsTouched();

    this.logger.info('Form state:', {
      valid: this.registerForm.valid,
      value: this.registerForm.value,
      errors: this.registerForm.errors,
    });

    // Check if form is valid
    if (!this.registerForm.valid) {
      this.logger.info('Form invalid, stopping submission');
      this.submitError.set('Please correct the errors in the form');
      return;
    }

    this.isLoading.set(true);
    this.submitError.set('');

    const formValue = this.registerForm.getRawValue();
    const registerData: RegisterAPIPayload = {
      company: formValue.company.trim(),
      lastname: formValue.lastname.trim(),
      firstname: formValue.firstname.trim(),
      email: formValue.email.toLowerCase().trim(),
      password: formValue.password,
    };

    try {
      this.logger.info('Calling registration API...', { ...registerData, password: '[REDACTED]' });
      const response: AuthResponse = await this.authService.callRegisterAPI(registerData);
      this.logger.info('API response received:', response);

      // Redirect to login page or dashboard
      await this.router.navigate(['/login']);
    } catch (error) {
      this.logger.error('Error during registration:', error);

      if (error instanceof HttpErrorResponse) {
        switch (error.status) {
          case 400:
            this.submitError.set('Invalid data. Please check your information.');
            break;
          case 409:
            this.submitError.set('An account with this email already exists.');
            break;
          case 500:
            this.submitError.set('Server error. Please try again later.');
            break;
          default:
            this.submitError.set('An error occurred. Please try again.');
        }
      } else {
        this.submitError.set('An unexpected error occurred.');
      }
    } finally {
      this.isLoading.set(false);
    }
  }
}