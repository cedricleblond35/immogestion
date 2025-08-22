import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PropertyExpenses } from './property-expenses';

describe('PropertyExpenses', () => {
  let component: PropertyExpenses;
  let fixture: ComponentFixture<PropertyExpenses>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyExpenses]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PropertyExpenses);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
