import { formatCurrency } from '@angular/common';
import { Component, viewChild  } from '@angular/core';
import { NgForm, FormsModule } from '@angular/forms';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './login.html',
  styleUrl: './login.scss',
  // template: `<br><br><br><br>
  //   <form #loginForm="ngForm" (ngSubmit)="onSubmit(loginForm)">
  //     <input type="text" name="username" ngModel required>
  //     <input type="password" name="password" ngModel required>
      
  //     <!-- Désactive le bouton si le formulaire n'est pas valide -->
  //     <button type="submit" [disabled]="!loginForm.valid">
  //       Login
  //     </button>
  //   </form>
  // `
})
export class Login {
  
  loginForm = viewChild<NgForm>('loginForm');

  autoSave() {
    const form = this.loginForm();

    if (form && form.dirty) {
      console.log('Auto-saving form data:', form.value);
    }
  }




  onSubmit() {
     const form = this.loginForm()
    if (form?.valid) {
      console.log(form.value); // { username: '...', password: '...' }
    }
  }
}