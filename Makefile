SUBDIRS = go node python
TARGETS = install build clean cleanall

all:

$(SUBDIRS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

$(TARGETS): $(SUBDIRS)

.PHONY: $(TARGETS) $(SUBDIRS)
