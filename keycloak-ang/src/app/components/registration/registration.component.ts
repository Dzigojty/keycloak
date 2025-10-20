import { Component } from '@angular/core';

@Component({
  selector: 'app-registration',
  templateUrl: './registration.component.html',
  styleUrls: ['./registration.component.css']
})
export class  RegistrationComponent {
  title = 'Регистрация в системе';
  
  // Добавьте эти свойства
  name = '';
  position = '';
  password1 = '';
  password2 = '';
  
  // Метод для обработки отправки формы
  onSubmit() {
    console.log('Name:', this.name);
    console.log('Position:', this.position);
    console.log('Password1:', this.password1);
    console.log('Password2:', this.password2);
    // Здесь будет логика создания пользователя
  }
}