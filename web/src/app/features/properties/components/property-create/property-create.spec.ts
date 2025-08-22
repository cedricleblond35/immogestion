import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PropertyCreate } from './property-create';

describe('PropertyCreate', () => {
  let component: PropertyCreate;
  let fixture: ComponentFixture<PropertyCreate>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyCreate]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PropertyCreate);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
