package main

/*
#include <dlfcn.h>
#include <objc/runtime.h>
#include <stdio.h>

static int (*getClassList)(void*,int) = NULL;
static char* (*getClassName)(Class) = NULL;
static Method* (*copyMethodList)(Class,int*) = NULL;
static SEL (*getSelector)(Method) = NULL;
static char* (*getSelName)(SEL) = NULL;
static Class (*objectGetClass)(void*) = NULL;

static void * open_handle(void)
{
	void * handle = dlopen("/usr/lib/libobjc.A.dylib",
		RTLD_LAZY | RTLD_GLOBAL | RTLD_NOLOAD);
	getClassList = dlsym(handle, "objc_getClassList");
	getClassName = dlsym(handle, "class_getName");
	copyMethodList = dlsym(handle, "class_copyMethodList");
	getSelector = dlsym(handle, "method_getName");
	getSelName = dlsym(handle, "sel_getName");
	objectGetClass = dlsym(handle, "object_getClass");

	return handle;
}

int do_getClassList(Class * ptr, int count)
{
	return getClassList(ptr, count);
}

char * do_getClassName(Class klass)
{
	return getClassName(klass);
}

Method * do_copyMethodList(Class klass, int * count)
{
	Method * methods = (Method*)malloc(sizeof(Method)*(*count));
	methods = copyMethodList(klass, count);
	return methods;
}

char * do_getSelName(Method method) {
	return getSelName(getSelector(method));
}

Class do_objectGetClass(void * klass)
{
	return objectGetClass(klass);
}

static Class * alloc_classes(int count)
{
	Class * classes = (Class*)malloc(sizeof(Class)*count);
	return classes;
}

*/
import "C"
import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type Method struct {
	handle   C.Method
	selector string
}

type Class struct {
	name            string
	handle          C.Class
	instanceMethods []Method
	classMethods    []Method
}

type Resolver struct {
	classCount int
	handle     unsafe.Pointer
	classes    []Class
}

func newResolver() *Resolver {
	res := &Resolver{}
	res.handle = C.open_handle()
	return res
}

func (r *Resolver) enumerateClasses() {
	r.classCount = int(C.do_getClassList(nil, 0))

	classes := C.alloc_classes(C.int(r.classCount))
	C.do_getClassList(classes, C.int(r.classCount))

	var cls []C.Class

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&cls))
	hdr.Cap = r.classCount
	hdr.Len = r.classCount
	hdr.Data = uintptr(unsafe.Pointer(classes))

	for _, class := range cls {
		name := C.GoString(C.do_getClassName(class))
		klass := Class{
			name:   name,
			handle: class,
		}
		var count C.int
		instanceMethods := C.do_copyMethodList(class, &count)

		var methods []C.Method

		methodsHdr := (*reflect.SliceHeader)(unsafe.Pointer(&methods))
		methodsHdr.Cap = int(count)
		methodsHdr.Len = int(count)
		methodsHdr.Data = uintptr(unsafe.Pointer(instanceMethods))

		for _, method := range methods {
			klass.instanceMethods = append(klass.instanceMethods, Method{
				handle:   method,
				selector: C.GoString(C.do_getSelName(method)),
			})
		}

		classMethods := C.do_copyMethodList(C.do_objectGetClass(unsafe.Pointer(class)), &count)

		var clsMethods []C.Method

		cMethodsHdr := (*reflect.SliceHeader)(unsafe.Pointer(&clsMethods))
		cMethodsHdr.Cap = int(count)
		cMethodsHdr.Len = int(count)
		cMethodsHdr.Data = uintptr(unsafe.Pointer(classMethods))

		for _, method := range clsMethods {
			klass.classMethods = append(klass.classMethods, Method{
				handle:   method,
				selector: C.GoString(C.do_getSelName(method)),
			})
		}

		r.classes = append(r.classes, klass)
	}
}

func (r *Resolver) printClasses() {
	for _, class := range r.classes {
		for _, method := range class.classMethods {
			fmt.Printf("+[%s %s]\n", class.name, method.selector)
		}
		for _, method := range class.instanceMethods {
			fmt.Printf("-[%s %s]\n", class.name, method.selector)
		}
	}
}

func (r *Resolver) getClass(className string) *Class {
	for _, class := range r.classes {
		if class.name == className {
			return &class
		}
	}
	return nil
}

func (r *Resolver) classContains(className string) []string {
	var classes []string
	for _, class := range r.classes {
		if strings.Contains(class.name, className) {
			classes = append(classes, class.name)
		}
	}
	return classes
}

func (r *Resolver) getAllClasses() []string {
	var classes []string
	for _, class := range r.classes {
		classes = append(classes, class.name)
	}
	return classes
}
