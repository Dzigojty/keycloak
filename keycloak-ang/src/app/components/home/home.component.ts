// components/home/home.component.ts
import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-home',
  template: `
    <div style="text-align: center; padding: 50px;">
      <h1>Добро пожаловать!</h1>
      <p>Keycloak инициализирован успешно</p>
      
      <div style="margin-top: 30px;">
        <button (click)="goToLogin()" 
                style="margin: 10px; padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px;">
          Вход
        </button>
        
        <button (click)="goToRegister()" 
                style="margin: 10px; padding: 10px 20px; background: #6c757d; color: white; border: none; border-radius: 4px;">
          Регистрация
        </button>
      </div>

      <div style="margin-top: 20px; color: green;">
        Кнопки должны работать теперь!
      </div>
    </div>
  `
})
export class HomeComponent {
  
  constructor(private router: Router) {}

  goToLogin() {
    console.log('Переход на страницу входа');
    this.router.navigate(['/login']);
  }

  goToRegister() {
    console.log('Переход на страницу регистрации');
    this.router.navigate(['/register']);
  }
}