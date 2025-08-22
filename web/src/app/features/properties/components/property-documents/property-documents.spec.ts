import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PropertyDocuments } from './property-documents';

describe('PropertyDocuments', () => {
  let component: PropertyDocuments;
  let fixture: ComponentFixture<PropertyDocuments>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PropertyDocuments]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PropertyDocuments);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
