import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PropertyTenants } from './property-tenants';

describe('PropertyTenants', () => {
  let component: PropertyTenants;
  let fixture: ComponentFixture<PropertyTenants>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyTenants]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PropertyTenants);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
