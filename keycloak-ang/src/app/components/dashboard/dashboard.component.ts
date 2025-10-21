// dashboard.component.ts
import { Component, OnInit } from '@angular/core';
import { KeycloakService } from 'keycloak-angular';

@Component({
  selector: 'app-dashboard',
  template: `
    <div class="dashboard">
      <h2>Добро пожаловать в Dashboard!</h2>
      <p>Вы успешно вошли через Keycloak</p>
      <button (click)="logout()" class="btn btn-danger">Выйти</button>
    </div>
  `
})
export class DashboardComponent implements OnInit {
  
  constructor(private keycloakService: KeycloakService) {}

  ngOnInit() {
    // Можно получить информацию о пользователе
    this.keycloakService.getKeycloakInstance().loadUserInfo().then(userInfo => {
      console.log('User info:', userInfo);
    });
  }

  logout() {
    this.keycloakService.logout();
  }
}