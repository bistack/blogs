int maxi(int i1, int i2)
{
  return (i1 > i2) ? i1 : i2;
}

#include <CUnit/CUnit.h>

void test_maxi(void)
{
  CU_ASSERT(maxi(0,2) == 2);
  CU_ASSERT(maxi(0,-2) == 0);
  CU_ASSERT(maxi(2,2) == 2);
}

int main(void)
{
  CU_ErrorCode err = CU_initialize_registry();

  CU_pSuite pSuit = CU_add_suite("testSuit1", NULL, NULL);

  CU_ADD_TEST(pSuit, test_maxi);

#include <CUnit/Basic.h>
  err = CU_basic_run_tests();

  CU_cleanup_registry();

  return err;
}
