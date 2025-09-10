export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterCredentials {
  formId: string;
  company: string;
  lastname: string;
  firstname: string;
  email: string;
  password: string;
  confirmPassword: string;
  terms: boolean;
}

export interface RegisterAPIPayload {
  company: string;
  lastname: string;
  firstname: string;
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