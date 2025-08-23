import { TestBed } from '@angular/core/testing';

import { AuthFeature } from './auth-feature';

describe('AuthFeature', () => {
  let service: AuthFeature;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(AuthFeature);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
