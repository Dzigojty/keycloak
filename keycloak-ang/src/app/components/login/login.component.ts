// login.component.ts
import { Component } from '@angular/core';
import { KeycloakService } from 'keycloak-angular';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html'
})
export class LoginComponent {
  name: string = '';
  password: string = '';

  constructor(private keycloakService: KeycloakService) {}

  // Метод для входа через Keycloak
  loginWithKeycloak() {
    this.keycloakService.login({
      redirectUri: window.location.origin + '/dashboard' // Куда редиректить после успешного входа
    });
  }

  // Традиционный вход (если нужен)
  onSubmit() {
    console.log('Логин:', this.name);
    console.log('Пароль:', this.password);
    // Здесь можно добавить кастомную логику входа
  }
}