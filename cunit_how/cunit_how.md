A Step by Step CUnit
==

测试目标: test target func
--

```C
int maxi(int i1, int i2)
{
  return (i1 > i2) ? i1 : i2;
}
```

1, 编写test case: write test case
--

test func函数签名：**void test_func(void)**

```C
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
```

2，初始化test register: initialize test register
--

```C
#include CUnit/CUnit.h

CU_ErrorCode CU_initialize_registry(void)
```

```C
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

  return err;
}
```

3, 添加test suite: add test suite
--

```C
#include CUnit/CUnit.h

CU_pSuite CU_add_suite(const char* strName, \
	CU_InitializeFunc pInit, CU_CleanupFunc pClean)
```


```C
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

  return err;
}
```


4, 添加test case: add test case
--


```C
#include CUnit/CUnit.h

CU_pTest CU_add_test(CU_pSuite pSuite, \
	const char* strName, CU_TestFunc pTestFunc)

#define CU_ADD_TEST(suite, test) \
	(CU_add_test(suite, #test, (CU_TestFunc)test))
```

```C
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

  return err;
}
```

5，调用test接口: run tests
--

```C
#include <CUnit/Automated.h>
void CU_automated_run_tests(void)		// 批量测试，结果输出到xml文件

#include <CUnit/Basic.h>
CU_ErrorCode CU_basic_run_tests(void)	// 批量测试，结果输出到标准输出

#include <CUnit/Console.h>
void CU_console_run_tests(void)			// 终端交互式测试

#include <CUnit/CUCurses.h>
void CU_curses_run_tests(void)			// curses交互式测试

```

```C
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

  return err;
}
```

6，清理test register: clean test register
--

```C
void CU_cleanup_registry(void)
```

```C
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
```

7，编译运行: compile & run
--

编译+动态链接

```sh
$ gcc -o maxi -lcunit maxi.c

$ ldd maxi
	linux-vdso.so.1 (0x00007ffc805e0000)
	libcunit.so.1 => /usr/lib/libcunit.so.1 (0x00007f7365d0c000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f7365961000)
	/lib64/ld-linux-x86-64.so.2 (0x000055b8a92aa000)

$ file maxi
maxi: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), \
dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, \
for GNU/Linux 2.6.32, \
BuildID[sha1]=235f7813d09dd11a8389826010e0b8be33869ff1, not stripped

$ ./maxi

     CUnit - A unit testing framework for C - Version 2.1-3
     http://cunit.sourceforge.net/



Run Summary:    Type  Total    Ran Passed Failed Inactive
              suites      1      1    n/a      0        0
               tests      1      1      1      0        0
             asserts      3      3      3      0      n/a

Elapsed time =    0.000 seconds
```

编译+静态链接

```sh
$ gcc -c maxi.c
$ gcc -o maxi maxi.o -static -lcunit

$ ldd maxi
	not a dynamic executable

$ file maxi
maxi: ELF 64-bit LSB executable, x86-64, version 1 (GNU/Linux), \
statically linked, \
for GNU/Linux 2.6.32, \
BuildID[sha1]=f32f4780e438b73aae39174c81ad54caad372cdd, not stripped

$ ./maxi


     CUnit - A unit testing framework for C - Version 2.1-2
     http://cunit.sourceforge.net/



Run Summary:    Type  Total    Ran Passed Failed Inactive
              suites      1      1    n/a      0        0
               tests      1      1      1      0        0
             asserts      3      3      3      0      n/a

Elapsed time =    0.000 seconds

```



