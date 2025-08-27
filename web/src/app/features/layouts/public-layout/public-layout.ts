import { Component } from '@angular/core';
import { RouterModule } from '@angular/router'; // <-- important

@Component({
  selector: 'app-public-layout',
  imports: [
    RouterModule  // <-- nécessaire pour router-outlet
  ],     
  templateUrl: './public-layout.html',
  styleUrl: './public-layout.scss'
})
export class PublicLayout {

}
