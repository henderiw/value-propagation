# value-propagation

This repo has implementation of value propagation using CEL or gotemplating.
The cel based approach has 2 implementation:
1. one that uses the complete template as a cel expression. (issue is that there is no way to differ between a string and cel variable/function)
2. one that parses the json data and valudates if the data is a string and provides the cellexpression at this level